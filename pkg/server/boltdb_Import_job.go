package server

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mkorenkov/covid-19/pkg/documents"
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/pkg/errors"
)

const importChanSize = 50
const batchSize = 1000

type sentinelError string

func (e sentinelError) Error() string {
	return string(e)
}

const ImportCancelledError = sentinelError("Context has been cancelled during import")

// due to data structure changes, a few weeks worths of data have
// "total_tests" stored in "tests_per_1m" json field. Important to note,
// "region" filed contains "population" of the country.
type legacyCountryData struct {
	Name          string    `json:"name"`
	When          time.Time `json:"when"`
	Cases         uint64    `json:"total_cases"`
	Deaths        uint64    `json:"total_deaths"`
	Tests         uint64    `json:"total_tests"`
	PossibleTests uint64    `json:"tests_per_1m"`
}

func parseCountry(payload []byte) (documents.DataEntry, error) {
	var res documents.DataEntry
	legacyCountryEntry := legacyCountryData{}
	if legacyParseErr := json.Unmarshal(payload, &legacyCountryEntry); legacyParseErr == nil && legacyCountryEntry.PossibleTests > legacyCountryEntry.Tests {
		return documents.DataEntry{
			Name:   legacyCountryEntry.Name,
			When:   legacyCountryEntry.When,
			Cases:  legacyCountryEntry.Cases,
			Deaths: legacyCountryEntry.Deaths,
			Tests:  legacyCountryEntry.PossibleTests,
		}, nil
	}
	if jsonErr := json.Unmarshal(payload, &res); jsonErr != nil {
		return res, errors.Wrap(jsonErr, "error decoding json from DB")
	}
	return res, nil
}

type importPayload struct {
	DataItem   documents.DataEntry
	Collection string
}

func readStates(ctx context.Context, wg *sync.WaitGroup, importDB *bolt.DB, data chan<- importPayload, errorChan chan<- error) {
	defer wg.Done()
	err := importDB.View(func(tx *bolt.Tx) error {
		masterCollectionBucket := tx.Bucket([]byte(documents.StateCollection))
		statesCursor := masterCollectionBucket.Cursor()

		for stateBucketKey, stateBucketName := statesCursor.First(); stateBucketKey != nil; stateBucketKey, stateBucketName = statesCursor.Next() {
			bucket := tx.Bucket([]byte(stateBucketName))
			if bucket == nil {
				return errors.Errorf("Bucket %s not found in import", stateBucketName)
			}

			c := bucket.Cursor()
			for key, payload := c.First(); key != nil; key, payload = c.Next() {
				dataEntry := documents.DataEntry{}
				if jsonErr := json.Unmarshal(payload, &dataEntry); jsonErr != nil {
					return errors.Wrap(jsonErr, "error decoding json from DB")
				}
				select {
				case <-ctx.Done():
					return errors.Wrap(ImportCancelledError, "States import got interrupted")
				case data <- importPayload{dataEntry, documents.StateCollection}:
				}
			}
		}
		return nil
	})
	if err != nil {
		errorChan <- errors.Wrap(err, "Error importing states")
		return
	}
}

func readCountries(ctx context.Context, wg *sync.WaitGroup, importDB *bolt.DB, data chan<- importPayload, errorChan chan<- error) {
	defer wg.Done()
	err := importDB.View(func(tx *bolt.Tx) error {
		masterCollectionBucket := tx.Bucket([]byte(documents.CountryCollection))
		countriesCursor := masterCollectionBucket.Cursor()

		for countryBucketKey, countryBucketName := countriesCursor.First(); countryBucketKey != nil; countryBucketKey, countryBucketName = countriesCursor.Next() {
			bucket := tx.Bucket([]byte(countryBucketName))
			if bucket == nil {
				return errors.Errorf("Bucket %s not found in import", countryBucketName)
			}

			c := bucket.Cursor()
			for key, payload := c.First(); key != nil; key, payload = c.Next() {
				dataEntry, parseErr := parseCountry(payload)
				if parseErr != nil {
					return errors.Wrap(parseErr, "error decoding country data")
				}
				select {
				case <-ctx.Done():
					return errors.Wrap(ImportCancelledError, "Countries import got interrupted")
				case data <- importPayload{dataEntry, documents.CountryCollection}:
				}
			}
		}
		return nil
	})
	if err != nil {
		errorChan <- errors.Wrap(err, "Error importing countries")
		return
	}
}

func batchImporter(ctx context.Context, wg *sync.WaitGroup, myDB *bolt.DB, importFromDB *bolt.DB, data <-chan importPayload, errorChan chan<- error) {
	defer wg.Done()

	statesBatch := make([]documents.CollectionEntry, 0, batchSize)
	countriesBatch := make([]documents.CollectionEntry, 0, batchSize)

BatcherLoop:
	for {
		select {
		case <-ctx.Done():
			errorChan <- errors.Wrap(ImportCancelledError, "Import got interrupted")
			return
		case payload, more := <-data:
			if !more {
				break BatcherLoop
			}
			switch payload.Collection {
			case documents.StateCollection:
				statesBatch = append(statesBatch, payload.DataItem)
				if len(statesBatch) >= batchSize {
					if iErr := documents.BulkSave(myDB, documents.StateCollection, statesBatch); iErr != nil {
						errorChan <- errors.Wrap(iErr, "Failed to import states")
						return
					}
					log.Printf("[DEBUG] imported %d state entries\n", len(statesBatch))
					statesBatch = make([]documents.CollectionEntry, 0, batchSize)
				}
			case documents.CountryCollection:
				countriesBatch = append(countriesBatch, payload.DataItem)
				if len(countriesBatch) >= batchSize {
					if iErr := documents.BulkSave(myDB, documents.CountryCollection, countriesBatch); iErr != nil {
						errorChan <- errors.Wrap(iErr, "Failed to import countries")
						return
					}
					log.Printf("[DEBUG] imported %d country entries\n", len(countriesBatch))
					countriesBatch = make([]documents.CollectionEntry, 0, batchSize)
				}
			}
		}
	}
	if len(statesBatch) >= 0 {
		if iErr := documents.BulkSave(myDB, documents.StateCollection, statesBatch); iErr != nil {
			errorChan <- errors.Wrap(iErr, "Failed to import states")
			return
		}
		log.Printf("[DEBUG] imported %d state entries\n", len(statesBatch))
	}
	if len(countriesBatch) >= 0 {
		if iErr := documents.BulkSave(myDB, documents.CountryCollection, countriesBatch); iErr != nil {
			errorChan <- errors.Wrap(iErr, "Failed to import countries")
			return
		}
		log.Printf("[DEBUG] imported %d country entries\n", len(countriesBatch))
	}
}

func importBoltDB(ctx context.Context, rctx *requestcontext.RequestContext, importDBPath string) {
	errorChan := rctx.Errors
	importDB, err := bolt.Open(importDBPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		errorChan <- errors.Wrapf(err, "Error opening import DB %s", importDBPath)
		return
	}
	defer func() {
		if cErr := importDB.Close(); cErr != nil {
			errorChan <- errors.Wrapf(cErr, "Error closing import DB %s", importDBPath)
			return
		}
		if cErr := os.Remove(importDBPath); cErr != nil {
			errorChan <- errors.Wrapf(cErr, "Error removing import DB %s", importDBPath)
		}
	}()

	importDataChan := make(chan importPayload, importChanSize)

	var writeWG sync.WaitGroup
	writeWG.Add(1)

	go batchImporter(ctx, &writeWG, rctx.DB, importDB, importDataChan, errorChan)

	var readerWG sync.WaitGroup
	readerWG.Add(2)

	go readCountries(ctx, &readerWG, importDB, importDataChan, errorChan)
	go readStates(ctx, &readerWG, importDB, importDataChan, errorChan)

	readerWG.Wait()

	close(importDataChan)

	writeWG.Wait()
}

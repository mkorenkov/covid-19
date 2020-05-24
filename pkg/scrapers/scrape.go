package scrapers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mkorenkov/covid-19/pkg/documents"
	"github.com/mkorenkov/covid-19/pkg/httpclient"
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/mkorenkov/covid-19/worldometers"
	"github.com/pkg/errors"
)

const (
	// StateCollection name of the states collection
	StateCollection = "States"
	// CountryCollection name of the countries collection
	CountryCollection = "Countries"
)

// Countries scrapes countries over some interval.
func Countries(ctx context.Context, interval time.Duration, backups chan<- documents.CollectionEntry) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	db := requestcontext.DB(ctx)
	if db == nil {
		panic(errors.New("Could not retrieve DB from context"))
	}
	errorChan := requestcontext.Errors(ctx)
	if errorChan == nil {
		panic(errors.New("Could not retrieve error chan from context"))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("[DEBUG] Scraping countries")
			countries, err := worldometers.Countries(ctx, httpclient.Retryable())
			if err != nil {
				errorChan <- errors.Wrap(err, "error scraping United States values")
			}
			err = db.Batch(func(tx *bolt.Tx) error {
				allCountriesBucket, txErr := tx.CreateBucketIfNotExists([]byte(CountryCollection))
				if txErr != nil {
					return errors.Wrapf(txErr, "error creating %s bucket", CountryCollection)
				}

				for _, country := range countries {
					if country.Name == "" {
						continue
					}

					countryDoc := documents.FromCountry(*country)
					backups <- countryDoc

					myCountryBucket, txErr := tx.CreateBucketIfNotExists([]byte(countryDoc.GetName()))
					if txErr != nil {
						return errors.Wrapf(txErr, "error creating %s bucket", countryDoc.GetName())
					}
					if txErr := allCountriesBucket.Put([]byte(countryDoc.GetName()), []byte(countryDoc.GetName())); txErr != nil {
						return errors.Wrapf(txErr, "error creating %s record in %s", countryDoc.GetName(), CountryCollection)
					}
					docBody, txErr := json.Marshal(countryDoc)
					if txErr != nil {
						return errors.Wrap(txErr, "JSON marshal error")
					}
					if txErr := myCountryBucket.Put([]byte(countryDoc.GetWhen().Format(time.RFC3339)), docBody); txErr != nil {
						return errors.Wrapf(txErr, "error creating %s record in %s", countryDoc.GetName(), CountryCollection)
					}
				}
				return nil
			})
			if err != nil {
				errorChan <- errors.Wrapf(err, "Error while writing %s data to DB", CountryCollection)
			}
			log.Printf("[INFO] Done scraping countries. Sleeping %s \n", interval)
		}
	}
}

// States scrapes states over some interval.
func States(ctx context.Context, interval time.Duration, backups chan<- documents.CollectionEntry) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	db := requestcontext.DB(ctx)
	if db == nil {
		panic(errors.New("Could not retrieve DB from context"))
	}
	errorChan := requestcontext.Errors(ctx)
	if errorChan == nil {
		panic(errors.New("Could not retrieve error chan from context"))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("[DEBUG] Scraping states")
			states, err := worldometers.States(ctx, httpclient.Retryable())
			if err != nil {
				errorChan <- errors.Wrap(err, "error scraping United States values")
			}
			err = db.Batch(func(tx *bolt.Tx) error {
				allStatesBucket, txErr := tx.CreateBucketIfNotExists([]byte(StateCollection))
				if txErr != nil {
					return errors.Wrapf(txErr, "error creating %s bucket", StateCollection)
				}

				for _, state := range states {
					if state.Name == "" {
						continue
					}

					stateDoc := documents.FromState(*state)
					backups <- stateDoc

					myCountryBucket, txErr := tx.CreateBucketIfNotExists([]byte(stateDoc.GetName()))
					if txErr != nil {
						return errors.Wrapf(txErr, "error creating %s bucket", stateDoc.GetName())
					}
					if txErr := allStatesBucket.Put([]byte(stateDoc.GetName()), []byte(stateDoc.GetName())); txErr != nil {
						return errors.Wrapf(txErr, "error creating %s record in %s", stateDoc.GetName(), StateCollection)
					}
					docBody, txErr := json.Marshal(stateDoc)
					if txErr != nil {
						return errors.Wrap(txErr, "JSON marshal error")
					}
					if txErr := myCountryBucket.Put([]byte(stateDoc.GetWhen().Format(time.RFC3339)), docBody); txErr != nil {
						return errors.Wrapf(txErr, "error creating %s record in %s", stateDoc.GetName(), StateCollection)
					}
				}
				return nil
			})
			if err != nil {
				errorChan <- errors.Wrapf(err, "Error while writing %s data to DB", StateCollection)
			}
			log.Printf("[INFO] Done scraping states. Sleeping %s \n", interval)
		}
	}
}

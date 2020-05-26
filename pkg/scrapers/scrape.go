package scrapers

import (
	"context"
	"log"
	"strings"
	"time"

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

	onTicker := func() {
		log.Println("[DEBUG] Scraping countries")
		rawCountries, err := worldometers.Countries(ctx, httpclient.Retryable())
		if err != nil {
			errorChan <- errors.Wrap(err, "error scraping Countries values")
		}
		countryDocs := []documents.CollectionEntry{}
		for _, country := range rawCountries {
			if country.Name == "" {
				continue
			}
			countryDoc := documents.FromCountry(*country)
			backups <- countryDoc
			countryDocs = append(countryDocs, *countryDoc)
		}
		err = documents.BulkSave(db, CountryCollection, countryDocs)
		if err != nil {
			errorChan <- errors.Wrapf(err, "Error while writing %s data to DB", CountryCollection)
		}
		log.Printf("[INFO] Done scraping countries. Sleeping %s \n", interval)
	}

	// force the first run on the app start
	onTicker()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			onTicker()
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

	onTicker := func() {
		log.Println("[DEBUG] Scraping states")
		rawStates, err := worldometers.States(ctx, httpclient.Retryable())
		if err != nil {
			errorChan <- errors.Wrap(err, "error scraping United States values")
		}
		statesDocs := []documents.CollectionEntry{}
		for _, state := range rawStates {
			if state.Name == "" {
				continue
			}
			stateDoc := documents.FromState(*state)
			backups <- stateDoc
			statesDocs = append(statesDocs, *stateDoc)
		}
		err = documents.BulkSave(db, StateCollection, statesDocs)
		if err != nil {
			errorChan <- errors.Wrapf(err, "Error while writing %s data to DB", StateCollection)
		}
		log.Printf("[INFO] Done scraping states. Sleeping %s \n", interval)
	}

	// force the first run on the app start
	onTicker()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			onTicker()
		}
	}
}

func key(doc documents.CollectionEntry) string {
	name := strings.TrimSpace(doc.GetName())
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, ". ", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

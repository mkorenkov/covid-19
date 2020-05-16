package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/mkorenkov/covid-19/worldometers"
	"github.com/pkg/errors"
)

const (
	envStorageDir            = "STORAGE_DIR"
	defaultStorageDir        = "/srv/data/covid-19"
	envScrapeIntervalSeconds = "SCRAPE_INTERVAL_SECONDS"
	defaultScrapeInterval    = 3 * time.Hour
)

func baseDir() string {
	// TODO: expand path
	if res := os.Getenv(envStorageDir); res != "" {
		return res
	}
	return defaultStorageDir
}

func scrapeInterval() time.Duration {
	if res := os.Getenv(envScrapeIntervalSeconds); res != "" {
		if val, err := strconv.ParseFloat(res, 64); err == nil {
			if val < 1800 {
				log.Fatalf("let's not query worldometers too often, %fs < 30m", val)
			}

			return time.Duration(val) * time.Second
		}
		log.Printf("failed to parse SCRAPE_INTERVAL_HOURS, defaulting to %v\n", defaultScrapeInterval)
	}
	return defaultScrapeInterval
}

func fileName(countryOrState string) string {
	res := strings.ToLower(countryOrState)
	res = strings.ReplaceAll(res, ". ", "_")
	res = strings.ReplaceAll(res, " ", "_")
	res = strings.ReplaceAll(res, ".", "_")
	res = strings.ReplaceAll(res, ":", "_")
	return fmt.Sprintf("%s.json", res)
}

func jsonDump(targetDir string, name string, payload interface{}) error {
	fullName := path.Join(targetDir, fileName(name))
	f, err := os.Create(fullName)
	if err != nil {
		return errors.Wrapf(err, "error opening %s", fullName)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(payload); err != nil {
		return errors.Wrapf(err, "error json encoding %s", name)
	}
	return nil
}

func scrapeData(ctx context.Context) error {
	targetDir := path.Join(baseDir(), time.Now().Format(time.RFC3339))
	_, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(targetDir, 0755)
		if errDir != nil {
			return errors.Wrapf(errDir, "failed to create %s", targetDir)
		}
	}

	countries, err := worldometers.Countries(ctx)
	if err != nil {
		return errors.Wrap(err, "error scraping per country values")
	}
	states, err := worldometers.States(ctx)
	if err != nil {
		return errors.Wrap(err, "error scraping United States values")
	}
	for name, payload := range countries {
		if err := jsonDump(targetDir, name, payload); err != nil {
			return errors.Wrapf(err, "failed to save JSON file %s", name)
		}
	}
	for name, payload := range states {
		if err := jsonDump(targetDir, name, payload); err != nil {
			return errors.Wrapf(err, "failed to save JSON file %s", name)
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(scrapeInterval())
	defer ticker.Stop()

	onTimer := func(ctx context.Context) {
		log.Println("Scraping")
		err := scrapeData(ctx)
		if err != nil {
			// TODO: log errors to external system
			log.Println(err)
		}
	}

	// force first call before
	onTimer(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			onTimer(ctx)
		}
	}
}

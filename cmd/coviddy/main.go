package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/logutils"
	"github.com/kelseyhightower/envconfig"
	"github.com/mkorenkov/covid-19/pkg/backup"
	"github.com/mkorenkov/covid-19/pkg/config"
	"github.com/mkorenkov/covid-19/pkg/documents"
	"github.com/mkorenkov/covid-19/pkg/reporter"
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/mkorenkov/covid-19/pkg/scrapers"
	"github.com/mkorenkov/covid-19/pkg/server"
	"github.com/pkg/errors"
)

const (
	dbName         = "coviddy.db"
	backupChanSize = 512
)

func init() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel("DEBUG"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
}

func ensureExists(dirName string) error {
	_, err := os.Stat(dirName)
	if !os.IsNotExist(err) {
		return errors.Wrap(err, "Unexpected error while calling os.Stat")
	}
	if derr := os.MkdirAll(dirName, 0755); derr != nil {
		return errors.Wrapf(derr, "failed to create %s", dirName)
	}
	return nil
}

func main() {
	var cfg config.Config
	if err := envconfig.Process("coviddy", &cfg); err != nil {
		log.Fatal(err)
	}
	if err := ensureExists(cfg.StorageDir); err != nil {
		log.Fatal(err)
	}
	if err := ensureExists(cfg.ImportsDir()); err != nil {
		log.Fatal(err)
	}
	if err := reporter.Initialize(cfg.SentryDSN); err != nil {
		log.Fatal(err)
	}

	myDB, err := bolt.Open(path.Join(cfg.StorageDir, dbName), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		cErr := myDB.Close()
		if cErr != nil {
			reporter.Report(cErr)
		}
	}()

	backupChan := make(chan documents.CollectionEntry, backupChanSize)
	defer close(backupChan)
	errorsChan := make(chan error)
	defer close(errorsChan)

	rctx := requestcontext.New(cfg, myDB, errorsChan, backupChan)
	ctx := requestcontext.WithContext(context.Background(), rctx)

	go reporter.ErrorReportingRoutine(errorsChan)
	go scrapers.States(ctx, cfg.ScrapeInterval, backupChan)
	go scrapers.Countries(ctx, cfg.ScrapeInterval, backupChan)
	go backup.ToS3(ctx, cfg, backupChan)

	b := server.NewBasicAuthMiddleware(cfg.Credentials)

	r := mux.NewRouter()
	r.HandleFunc("/", server.HomeHandler)

	internal := r.PathPrefix("/api/internal/v1/").Subrouter()
	internal.HandleFunc("/countries", server.UpsertAnythingHandler).Methods("POST")
	internal.HandleFunc("/states", server.UpsertAnythingHandler).Methods("POST")
	internal.HandleFunc("/import/country_or_state", server.UpsertAnythingHandler).Methods("POST")
	internal.HandleFunc("/boltdb/import", server.BoltDBImportHandler).Methods("POST")
	internal.Use(b.BasicAuth)

	api := r.PathPrefix("/api/v1/").Subrouter()
	api.HandleFunc("/countries", server.ListCountriesHandler).Methods("GET")
	api.HandleFunc("/states", server.ListStatesHandler).Methods("GET")
	api.HandleFunc("/countries/{country}", server.CountryDatapointsHandler).Methods("GET")
	api.HandleFunc("/states/{state}", server.StateDatapointsHandler).Methods("GET")

	log.Printf("[INFO] Listening %s\n", cfg.ListenAddr)

	srv := &http.Server{
		Handler:      server.PanicRecoveryMiddleware(server.LogMiddleware(requestcontext.InjectRequestContextMiddleware(r, rctx))),
		Addr:         cfg.ListenAddr,
		WriteTimeout: 1 * time.Minute,
		ReadTimeout:  1 * time.Minute,
	}
	log.Fatal(srv.ListenAndServe())
}

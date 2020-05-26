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
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/mkorenkov/covid-19/pkg/scrapers"
	"github.com/mkorenkov/covid-19/pkg/server"
	"github.com/pkg/errors"
)

const dbName = "covid19.db"

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
	if err := envconfig.Process("covid19", &cfg); err != nil {
		log.Fatal(err)
	}
	if err := ensureExists(cfg.StorageDir); err != nil {
		log.Fatal(err)
	}

	myDB, err := bolt.Open(path.Join(cfg.StorageDir, dbName), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	backupChan := make(chan documents.CollectionEntry)
	defer close(backupChan)
	errorsChan := make(chan error)
	defer close(errorsChan)

	rctx := requestcontext.New(myDB, errorsChan)
	ctx := requestcontext.WithContext(context.Background(), rctx)

	go scrapers.States(ctx, cfg.ScrapeInterval, backupChan)
	go scrapers.Countries(ctx, cfg.ScrapeInterval, backupChan)
	go backup.ToS3(ctx, cfg, backupChan)

	b := server.NewBasicAuthMiddleware(cfg.Credentials)

	r := mux.NewRouter()
	r.HandleFunc("/", server.HomeHandler)

	internal := r.PathPrefix("/api/internal/v1/").Subrouter()
	internal.HandleFunc("/countries", server.HomeHandler).Methods("POST")
	internal.HandleFunc("/states", server.HomeHandler).Methods("POST")
	internal.Use(b.BasicAuth)

	api := r.PathPrefix("/api/v1/").Subrouter()
	api.HandleFunc("/countries", server.HomeHandler).Methods("GET")
	api.HandleFunc("/states", server.HomeHandler).Methods("GET")
	api.HandleFunc("/countries/{country}", server.CountryDatapointsHandler).Methods("GET")
	api.HandleFunc("/states/{state}", server.StateDatapointsHandler).Methods("GET")

	log.Printf("[INFO] Listening %s\n", cfg.ListenAddr)

	srv := &http.Server{
		Handler:      server.PanicRecoveryMiddleware(server.LogMiddleware(requestcontext.InjectRequestContextMiddleware(r, rctx))),
		Addr:         cfg.ListenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/HouzuoGuo/tiedot/db"
// 	"github.com/HouzuoGuo/tiedot/dberr"
// 	"github.com/gorilla/mux"
// 	"github.com/hashicorp/logutils"
// 	"github.com/mkorenkov/covid-19/requestcontext"
// 	"github.com/mkorenkov/covid-19/worldometers"
// 	"github.com/pkg/errors"
// )

// const (
// 	envStorageDir            = "DB_DIR"
// 	envBackupDir            = "BACKUP_DIR"
// 	envScrapeIntervalSeconds = "SCRAPE_INTERVAL_SECONDS"
// 	defaultStorageDir        = "/tmp/covid-19/data"
// 	defaultBackupDir        = "/tmp/covid-19/backup"
// 	defaultScrapeInterval    = 3 * time.Hour
// )

// func init() {
// 	filter := &logutils.LevelFilter{
// 		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
// 		MinLevel: logutils.LogLevel("INFO"),
// 		Writer:   os.Stderr,
// 	}
// 	log.SetOutput(filter)
// }

// func dbDir() (string, error) {
// 	targetDir := defaultStorageDir
// 	if res := os.Getenv(envStorageDir); res != "" {
// 		targetDir = res
// 	}

// 	_, err := os.Stat(targetDir)
// 	if os.IsNotExist(err) {
// 		errDir := os.MkdirAll(targetDir, 0755)
// 		if errDir != nil {
// 			return "", errors.Wrapf(errDir, "failed to create %s", targetDir)
// 		}
// 	}
// 	return targetDir, nil
// }

// func backupDir() (string, error) {
// 	targetDir := defaultBackupDir
// 	if res := os.Getenv(envBackupDir); res != "" {
// 		targetDir = res
// 	}
// }

// func EmbeddedExample() {
// 	myDBDir, err := dbDir()
// 	if err != nil {
// 		panic(err)
// 	}
// 	myDB, err := db.OpenDB(myDBDir)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Create two collections: Feeds and Votes
// 	if err := myDB.Create("Feeds"); err != nil {
// 		panic(err)
// 	}
// 	if err := myDB.Create("Votes"); err != nil {
// 		panic(err)
// 	}

// 	// What collections do I now have?
// 	for _, name := range myDB.AllCols() {
// 		fmt.Printf("I have a collection called %s\n", name)
// 	}

// 	// Rename collection "Votes" to "Points"
// 	if err := myDB.Rename("Votes", "Points"); err != nil {
// 		panic(err)
// 	}

// 	// Drop (delete) collection "Points"
// 	if err := myDB.Drop("Points"); err != nil {
// 		panic(err)
// 	}

// 	// Scrub (repair and compact) "Feeds"
// 	if err := myDB.Scrub("Feeds"); err != nil {
// 		panic(err)
// 	}

// 	// ****************** Document Management ******************

// 	// Start using a collection (the reference is valid until DB schema changes or Scrub is carried out)
// 	feeds := myDB.Use("Feeds")

// 	// Insert document (afterwards the docID uniquely identifies the document and will never change)
// 	docID, err := feeds.Insert(map[string]interface{}{
// 		"name": "Go 1.2 is released",
// 		"url":  "golang.org"})
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Read document
// 	readBack, err := feeds.Read(docID)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Document", docID, "is", readBack)

// 	// Update document
// 	err = feeds.Update(docID, map[string]interface{}{
// 		"name": "Go is very popular",
// 		"url":  "google.com"})
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Process all documents (note that document order is undetermined)
// 	feeds.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
// 		fmt.Println("Document", id, "is", string(docContent))
// 		return true  // move on to the next document OR
// 		return false // do not move on to the next document
// 	})

// 	// Delete document
// 	if err := feeds.Delete(docID); err != nil {
// 		panic(err)
// 	}

// 	// More complicated error handing - identify the error Type.
// 	// In this example, the error code tells that the document no longer exists.
// 	if err := feeds.Delete(docID); dberr.Type(err) == dberr.ErrorNoDoc {
// 		fmt.Println("The document was already deleted")
// 	}

// 	// ****************** Index Management ******************
// 	// Indexes assist in many types of queries
// 	// Create index (path leads to document JSON attribute)
// 	if err := feeds.Index([]string{"author", "name", "first_name"}); err != nil {
// 		panic(err)
// 	}
// 	if err := feeds.Index([]string{"Title"}); err != nil {
// 		panic(err)
// 	}
// 	if err := feeds.Index([]string{"Source"}); err != nil {
// 		panic(err)
// 	}

// 	// What indexes do I have on collection A?
// 	for _, path := range feeds.AllIndexes() {
// 		fmt.Printf("I have an index on path %v\n", path)
// 	}

// 	// Remove index
// 	if err := feeds.Unindex([]string{"author", "name", "first_name"}); err != nil {
// 		panic(err)
// 	}

// 	// ****************** Queries ******************
// 	// Prepare some documents for the query
// 	feeds.Insert(map[string]interface{}{"Title": "New Go release", "Source": "golang.org", "Age": 3})
// 	feeds.Insert(map[string]interface{}{"Title": "Kitkat is here", "Source": "google.com", "Age": 2})
// 	feeds.Insert(map[string]interface{}{"Title": "Good Slackware", "Source": "slackware.com", "Age": 1})

// 	var query interface{}
// 	json.Unmarshal([]byte(`[{"eq": "New Go release", "in": ["Title"]}, {"eq": "slackware.com", "in": ["Source"]}]`), &query)

// 	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys

// 	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
// 		panic(err)
// 	}

// 	// Query result are document IDs
// 	for id := range queryResult {
// 		// To get query result document, simply read it
// 		readBack, err := feeds.Read(id)
// 		if err != nil {
// 			panic(err)
// 		}
// 		fmt.Printf("Query returned document %v\n", readBack)
// 	}

// 	// Gracefully close database
// 	if err := myDB.Close(); err != nil {
// 		panic(err)
// 	}
// }

// func scrapeInterval() time.Duration {
// 	if res := os.Getenv(envScrapeIntervalSeconds); res != "" {
// 		if val, err := strconv.ParseFloat(res, 64); err == nil {
// 			if val < 1800 {
// 				log.Fatalf("let's not query worldometers too often, %fs < 30m", val)
// 			}

// 			return time.Duration(val) * time.Second
// 		}
// 		log.Printf("failed to parse SCRAPE_INTERVAL_HOURS, defaulting to %v\n", defaultScrapeInterval)
// 	}
// 	return defaultScrapeInterval
// }

// func fileName(countryOrState string) string {
// 	res := strings.ToLower(countryOrState)
// 	res = strings.ReplaceAll(res, ". ", "_")
// 	res = strings.ReplaceAll(res, " ", "_")
// 	return fmt.Sprintf("%s.json", res)
// }

// func jsonDump(targetDir string, name string, payload interface{}) error {
// 	fullName := path.Join(targetDir, fileName(name))
// 	f, err := os.Create(fullName)
// 	if err != nil {
// 		return errors.Wrapf(err, "error opening %s", fullName)
// 	}
// 	defer f.Close()
// 	enc := json.NewEncoder(f)
// 	if err := enc.Encode(payload); err != nil {
// 		return errors.Wrapf(err, "error json encoding %s", name)
// 	}
// 	return nil
// }

// func scrape(ctx context.Context, interval time. countriesChan chan<-worldometers.Country, statesChan chan<-worldometers.UnitedState) {
// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()

// 	onTimerCountries := func(ctx context.Context) error {
// 		log.Println("[Debug] Scraping countries")
// 		countries, err := worldometers.Countries(ctx)
// 		if err != nil {
// 			return errors.Wrap(err, "error scraping per country values")
// 		}
// 		for _, country := range {
// 			countriesChan <- country
// 		}
// 		log.Println("[INFO] Done scraping countries")
// 	}

// 	onTimerStates := func(ctx context.Context) error {
// 		log.Println("[DEBUG] Scraping states")
// 		states, err := worldometers.States(ctx)
// 		if err != nil {
// 			return errors.Wrap(err, "error scraping United States values")
// 		}
// 		for _, country := range {
// 			countriesChan <- country
// 		}
// 		log.Println("[INFO] Done scraping states")
// 	}

// 	// force first call before
// 	onTimer(ctx)
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-ticker.C:
// 			onTimer(ctx)
// 		}
// 	}
// }

// func scrapeData(ctx context.Context) error {
// 	countries, err := worldometers.Countries(ctx)
// 	if err != nil {
// 		return errors.Wrap(err, "error scraping per country values")
// 	}
// 	states, err := worldometers.States(ctx)
// 	if err != nil {
// 		return errors.Wrap(err, "error scraping United States values")
// 	}
// 	for name, payload := range countries {
// 		if err := jsonDump(targetDir, name, payload); err != nil {
// 			return errors.Wrapf(err, "failed to save JSON file %s", name)
// 		}
// 	}
// 	for name, payload := range states {
// 		if err := jsonDump(targetDir, name, payload); err != nil {
// 			return errors.Wrapf(err, "failed to save JSON file %s", name)
// 		}
// 	}
// 	return nil
// }

// func dbConn(collections []string) (*db.DB, error) {
// 	myDBDir, err := dbDir()
// 	if err != nil {
// 		return nil, errors.Wrap(err, "unable to get db settings")
// 	}
// 	myDB, err := db.OpenDB(myDBDir)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "unable to create db conn")
// 	}
// 	for _, collection := range collections {
// 		if err := myDB.Create(collection); err != nil {
// 			myDB.Close()
// 			return nil, errors.Wrapf(err, "unable to create %s collection", collection)
// 		}
// 	}
// 	return myDB, nil
// }

// func main() {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	myDB, err := dbConn()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer myDB.Close()

// 	go ScrapeStates(ctx, scrapeInterval(), backupDir(), myDB)
// 	go ScrapeStates(ctx, scrapeInterval(), backupDir(), myDB)

// 	reqCtx := requestcontext.New(myDB)
// 	r := mux.NewRouter()
// 	r.HandleFunc("/", HomeHandler)
// 	srv := &http.Server{
// 		Handler:      requestcontext.Inject(r, reqCtx),
// 		Addr:         "127.0.0.1:8000",
// 		WriteTimeout: 15 * time.Second,
// 		ReadTimeout:  15 * time.Second,
// 	}
// 	log.Fatal(srv.ListenAndServe())

// }

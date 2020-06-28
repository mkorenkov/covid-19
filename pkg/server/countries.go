package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/mkorenkov/covid-19/pkg/documents"
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/pkg/errors"
)

const (
	beforeParam = "before"
	afterParam  = "after"
)

func writeError(w http.ResponseWriter, httpStatus int, msg string) {
	w.WriteHeader(httpStatus)
	w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, msg)))
}

// ListCountriesHandler prints per country data.
func ListCountriesHandler(w http.ResponseWriter, r *http.Request) {
	db := requestcontext.DB(r.Context())
	if db == nil {
		panic(errors.New("Could not retrieve DB from context"))
	}

	res := []string{}
	err := db.View(func(tx *bolt.Tx) error {
		masterCollectionBucket := tx.Bucket([]byte(documents.CountryCollection))
		c := masterCollectionBucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			res = append(res, string(v))
		}
		return nil
	})
	enc := json.NewEncoder(w)
	if err = enc.Encode(res); err != nil {
		panic(err)
	}
}

// UpsertCountriesHandler adds country to the DB.
func UpsertCountriesHandler(w http.ResponseWriter, r *http.Request) {
	db := requestcontext.DB(r.Context())
	if db == nil {
		panic(errors.New("Could not retrieve DB from context"))
	}
	payload, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		panic(errors.Wrap(readErr, "Error reading request body"))
	}
	dataEntry, parseErr := documents.Parse(payload)
	if parseErr != nil {
		if errors.Is(parseErr, documents.BucketNotFoundError) {
			http.Error(w, parseErr.Error(), http.StatusFailedDependency)
			w.WriteHeader(http.StatusCreated)
		}
		panic(parseErr)
	}
	if err := documents.Save(db, documents.CountryCollection, dataEntry); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)
}

// CountryDatapointsHandler prints per country data.
func CountryDatapointsHandler(w http.ResponseWriter, r *http.Request) {
	db := requestcontext.DB(r.Context())
	if db == nil {
		panic(errors.New("Could not retrieve DB from context"))
	}

	vars := mux.Vars(r)
	country := vars["country"]

	if country == "" {
		writeError(w, http.StatusBadRequest, "country param is required")
		return
	}

	res := map[string]documents.DataEntry{}

	min := r.URL.Query().Get(beforeParam)
	if min == "" {
		min = time.Time{}.In(time.Local).Format(time.RFC3339)
	}
	max := r.URL.Query().Get(afterParam)
	if max == "" {
		max = time.Now().Format(time.RFC3339)
	}

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(strings.ToLower(country)))
		if bucket == nil {
			writeError(w, http.StatusNotFound, "country not found")
			return nil
		}

		c := bucket.Cursor()
		for k, v := c.Seek([]byte(min)); k != nil && bytes.Compare(k, []byte(max)) <= 0; k, v = c.Next() {
			m := documents.DataEntry{}
			if jsonErr := json.Unmarshal(v, &m); jsonErr != nil {
				return errors.Wrap(jsonErr, "error decoding json from DB")
			}
			res[string(k)] = m
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err = enc.Encode(res); err != nil {
		panic(err)
	}
}

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		masterCollectionBucket, txErr := tx.CreateBucketIfNotExists([]byte(documents.CountryCollection))
		if txErr != nil {
			return errors.Wrapf(txErr, "error creating %s bucket", documents.CountryCollection)
		}
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
	m := documents.CountryEntry{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&m); err != nil {
		panic(err)
	}
	if err := documents.Save(db, documents.CountryCollection, m); err != nil {
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

	res := map[string]documents.CountryEntry{}

	min := r.URL.Query().Get(beforeParam)
	if min == "" {
		min = time.Time{}.Format(time.RFC3339)
	}
	max := r.URL.Query().Get(afterParam)
	if min == "" {
		min = time.Now().Format(time.RFC3339)
	}

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(strings.ToLower(country)))
		if bucket == nil {
			writeError(w, http.StatusNotFound, "country not found")
			return nil
		}

		c := bucket.Cursor()
		for k, v := c.Seek([]byte(min)); k != nil && bytes.Compare(k, []byte(max)) <= 0; k, v = c.Next() {
			m := documents.CountryEntry{}
			if jsonErr := json.Unmarshal(v, &m); jsonErr != nil {
				return errors.Wrap(jsonErr, "error decoding json from DB")
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(w)
	if err = enc.Encode(res); err != nil {
		panic(err)
	}
}

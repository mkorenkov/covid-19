package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/mkorenkov/covid-19/pkg/documents"
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/pkg/errors"
)

// ListStatesHandler prints per country data.
func ListStatesHandler(w http.ResponseWriter, r *http.Request) {
	db := requestcontext.DB(r.Context())
	if db == nil {
		panic(errors.New("Could not retrieve DB from context"))
	}

	res := []string{}
	err := db.View(func(tx *bolt.Tx) error {
		masterCollectionBucket := tx.Bucket([]byte(documents.StateCollection))
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

// UpsertStatesHandler adds country to the DB.
func UpsertStatesHandler(w http.ResponseWriter, r *http.Request) {
	db := requestcontext.DB(r.Context())
	if db == nil {
		panic(errors.New("Could not retrieve DB from context"))
	}
	m := documents.DataEntry{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&m); err != nil {
		panic(err)
	}
	if err := documents.Save(db, documents.StateCollection, m); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)
}

// StateDatapointsHandler prints per state data.
func StateDatapointsHandler(w http.ResponseWriter, r *http.Request) {
	db := requestcontext.DB(r.Context())
	if db == nil {
		panic(errors.New("Could not retrieve DB from context"))
	}

	vars := mux.Vars(r)
	state := vars["state"]

	if state == "" {
		writeError(w, http.StatusBadRequest, "state param is required")
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
		bucket := tx.Bucket([]byte(strings.ToLower(state)))
		if bucket == nil {
			writeError(w, http.StatusNotFound, "state not found")
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

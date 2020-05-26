package server

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
)

const (
	beforeParam = "before"
	afterParam  = "after"
)

func writeError(w http.ResponseWriter, httpStatus int, msg string) {
	w.WriteHeader(httpStatus)
	w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, msg)))
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

	min := r.URL.Query().Get(beforeParam)
	max := r.URL.Query().Get(afterParam)

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(strings.ToLower(country)))
		if bucket == nil {
			writeError(w, http.StatusNotFound, "country not found")
			return nil
		}

		c := bucket.Cursor()
		for k, v := c.Seek([]byte(min)); k != nil && bytes.Compare(k, []byte(max)) <= 0; k, v = c.Next() {
			fmt.Fprintf(w, "%v,", v)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

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

	min := r.URL.Query().Get(beforeParam)
	max := r.URL.Query().Get(afterParam)

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(strings.ToLower(state)))
		if bucket == nil {
			writeError(w, http.StatusNotFound, "state not found")
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

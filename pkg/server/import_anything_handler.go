package server

import (
	"io/ioutil"
	"net/http"

	"github.com/mkorenkov/covid-19/pkg/documents"
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/pkg/errors"
)

// UpsertAnythingHandler accepts boltdb, imports contents
func UpsertAnythingHandler(w http.ResponseWriter, r *http.Request) {
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
	s3Chan := requestcontext.ToS3(r.Context())
	if s3Chan == nil {
		panic(errors.New("Could not backup chan from context"))
	}
	s3Chan <- dataEntry
	if err := documents.FindBucketAndSave(db, dataEntry); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)
}

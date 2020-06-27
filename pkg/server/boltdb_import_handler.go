package server

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/pkg/errors"
)

type noCancel struct {
	ctx context.Context
}

func (c noCancel) Deadline() (time.Time, bool)       { return time.Time{}, false }
func (c noCancel) Done() <-chan struct{}             { return nil }
func (c noCancel) Err() error                        { return nil }
func (c noCancel) Value(key interface{}) interface{} { return c.ctx.Value(key) }

// WithoutCancel returns a context that is never canceled.
func WithoutCancel(ctx context.Context) context.Context {
	return noCancel{ctx: ctx}
}

// BoltDBImportHandler accepts boltdb, imports contents
func BoltDBImportHandler(w http.ResponseWriter, r *http.Request) {
	rc := requestcontext.GetRequestContext(r.Context())
	if rc == nil {
		panic(errors.New("Could not retrieve required context data from context"))
	}

	tmpDBFile, err := ioutil.TempFile(rc.Config.ImportsDir(), "bolt-import-*.db")
	if err != nil {
		panic(errors.Wrap(err, "Cannot create temp file for DB import"))
	}
	defer func() {
		tmpDBFile.Close()
	}()
	bufW := bufio.NewWriter(tmpDBFile)
	_, err = io.Copy(bufW, r.Body)
	if err != nil {
		panic(errors.Wrap(err, "Error writing to temp DB file"))
	}
	err = bufW.Flush()
	if err != nil {
		panic(errors.Wrap(err, "Error flushing buffer"))
	}
	err = tmpDBFile.Sync()
	if err != nil {
		panic(errors.Wrap(err, "Error fsyncing temp DB file"))
	}

	go importBoltDB(WithoutCancel(r.Context()), rc, tmpDBFile.Name())

	w.WriteHeader(http.StatusCreated)
}

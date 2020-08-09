package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mkorenkov/covid-19/pkg/documents"
	"github.com/mkorenkov/covid-19/pkg/httpclient"
	"github.com/pkg/errors"
)

const documentsChanSize = 32
const noopErr = sentinelError("Ignored (e.g. git file)")

type sentinelError string

func (e sentinelError) Error() string {
	return string(e)
}

type Config struct {
	ImportDir       string `split_words:"true" required:"true"`
	CoviddyURI      string `split_words:"true" required:"true"`
	CoviddyUser     string `split_words:"true" required:"true"`
	CoviddyPassword string `split_words:"true" required:"true"`
}

func parse(dateOverride time.Time, path string) (documents.CollectionEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening file %s", path)
	}
	defer f.Close()

	var reader io.Reader
	reader = f
	if strings.HasSuffix(path, ".gz") {
		zipReader, err := gzip.NewReader(f)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating gzip reader for %s", path)
		}
		defer zipReader.Close()
		reader = zipReader
	}
	allBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading from file %s", path)
	}
	doc, err := documents.NoValidationsParse(allBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing file %s", path)
	}
	doc.When = dateOverride
	return doc, nil
}

func getDateFromPath(path string) (time.Time, error) {
	switch {
	case strings.HasSuffix(path, ".json"):
		dir := filepath.Base(filepath.Dir(path))
		when, timeErr := time.Parse(time.RFC3339, dir)
		if timeErr != nil {
			return time.Time{}, errors.Wrapf(timeErr, "Error figuring out date/time for %s", path)
		}
		return when, nil
	case strings.HasSuffix(path, ".json.gz"):
		filename := strings.ReplaceAll(filepath.Base(path), ".json.gz", "")
		when, timeErr := time.Parse(time.RFC3339, filename)
		if timeErr != nil {
			return time.Time{}, errors.Wrapf(timeErr, "Error figuring out date/time for %s", path)
		}
		return when, nil
	case strings.Contains(path, "/.git/"):
		return time.Time{}, errors.Wrapf(noopErr, ".git in path %s", path)
	case strings.Contains(path, "/.hg/"):
		return time.Time{}, errors.Wrapf(noopErr, ".hg in path %s", path)
	case strings.Contains(path, "/.svn/"):
		return time.Time{}, errors.Wrapf(noopErr, ".svn in path %s", path)
	}
	return time.Time{}, errors.Errorf("unknown date/time use case for %s", path)
}

func parseAll(path string, out chan<- documents.CollectionEntry, errorChan chan<- error) {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.Mode().IsRegular() {
			when, timeErr := getDateFromPath(path)
			if timeErr != nil {
				if errors.Is(timeErr, noopErr) {
					return nil
				}
				return timeErr
			}
			doc, parseErr := parse(when, path)
			if parseErr != nil {
				return parseErr
			}
			out <- doc
		}
		return nil
	})
	if err != nil {
		errorChan <- err
	}
}

func onError(errorChan <-chan error) {
	for err := range errorChan {
		log.Fatal(err)
	}
}

func process(ctx context.Context, cfg Config, docs <-chan documents.CollectionEntry, errorChan chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		case doc, more := <-docs:
			if !more {
				return
			}
			if doc.GetName() != "" {
				if err := r(cfg, http.DefaultClient, doc); err != nil {
					errorChan <- errors.Wrapf(err, "Failed to upload %s", doc)
				}
			}
		}
	}
}

func r(cfg Config, httpClient httpclient.HTTPClient, doc documents.CollectionEntry) error {
	errChan := make(chan error)
	defer close(errChan)
	var buf bytes.Buffer

	if err := doc.Save(&buf); err != nil {
		return errors.Wrapf(err, "Error saving doc %s", doc)
	}
	req, err := http.NewRequest("POST", cfg.CoviddyURI, &buf)
	if err != nil {
		return errors.Wrap(err, "Error creating HTTP request")
	}
	req.SetBasicAuth(cfg.CoviddyUser, cfg.CoviddyPassword)
	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Error uploading doc %s", doc)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Error reading response body")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.Errorf("Unexpected HTTP status %d", resp.StatusCode)
	}

	return nil
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	docChan := make(chan documents.CollectionEntry, documentsChanSize)

	errorChan := make(chan error)
	defer close(errorChan)

	var wg sync.WaitGroup
	wg.Add(1)
	go onError(errorChan)
	go func() {
		defer close(docChan)
		parseAll(cfg.ImportDir, docChan, errorChan)
	}()
	go func() {
		defer wg.Done()
		process(context.Background(), cfg, docChan, errorChan)
	}()
	wg.Wait()
}

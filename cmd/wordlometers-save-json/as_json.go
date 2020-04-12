package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mkorenkov/covid-19-parser/worldometers"
	"github.com/pkg/errors"
)

func fileName(countryOrState string) string {
	res := strings.ToLower(countryOrState)
	res = strings.ReplaceAll(res, ". ", "_")
	res = strings.ReplaceAll(res, " ", "_")
	return fmt.Sprintf("%s.json", res)
}

func jsonDump(targetDir string, name string, payload interface{}) error {
	fullName := path.Join(targetDir, fileName(name))
	f, err := os.Create(fullName)
	if err != nil {
		return errors.Wrapf(err, "error opening %s", fullName)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(payload); err != nil {
		return errors.Wrapf(err, "error json encoding %s", name)
	}
	return nil
}

func main() {
	targetDir := path.Join("/tmp/json", time.Now().Format(time.RFC3339))
	_, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(targetDir, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	countries, err := worldometers.Countries()
	if err != nil {
		log.Fatal(err)
	}
	states, err := worldometers.States()
	if err != nil {
		log.Fatal(err)
	}
	for name, payload := range countries {
		err := jsonDump(targetDir, name, payload)
		if err != nil {
			log.Fatal(err)
		}
	}
	for name, payload := range states {
		err := jsonDump(targetDir, name, payload)
		if err != nil {
			log.Fatal(err)
		}
	}
}

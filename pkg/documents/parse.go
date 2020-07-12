package documents

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// DateRequiredError parsing error.
const DateRequiredError = sentinelError("DataEntry does not have a timestamp")

// DateRequiredError parsing error.
const InvalidDateError = sentinelError("DataEntry timestamp is earlier than Dec 31, 2019")

// due to data structure changes, a few weeks worths of data have
// "total_tests" stored in "tests_per_1m" json field. Important to note,
// "region" filed contains "population" of the country.
type legacyCountryData struct {
	Name           string    `json:"name"`
	When           time.Time `json:"when"`
	Cases          uint64    `json:"total_cases"`
	Deaths         uint64    `json:"total_deaths"`
	Tests          uint64    `json:"total_tests"`
	PossibleCases  uint64    `json:"cases_per_1m"`  // make sure to ignore when importing from https://github.com/edoc-hcraes/covid-19-data
	PossibleDeaths uint64    `json:"deaths_per_1m"` // make sure to ignore when importing from https://github.com/edoc-hcraes/covid-19-data
	PossibleTests  uint64    `json:"tests_per_1m"`  // except for corrupted entries: make sure to ignore when importing from https://github.com/edoc-hcraes/covid-19-data
}

func parse(payload []byte) (DataEntry, error) {
	var res DataEntry
	legacyCountryEntry := legacyCountryData{}
	if legacyParseErr := json.Unmarshal(payload, &legacyCountryEntry); legacyParseErr == nil {
		res = DataEntry{
			Name:   legacyCountryEntry.Name,
			When:   legacyCountryEntry.When,
			Cases:  legacyCountryEntry.Cases,
			Deaths: legacyCountryEntry.Deaths,
			Tests:  legacyCountryEntry.Tests,
		}
		if legacyCountryEntry.PossibleCases > legacyCountryEntry.Cases {
			res.Cases = legacyCountryEntry.PossibleCases
		}
		if legacyCountryEntry.PossibleDeaths > legacyCountryEntry.Deaths {
			res.Deaths = legacyCountryEntry.PossibleDeaths
		}
		if legacyCountryEntry.PossibleTests > legacyCountryEntry.Tests {
			res.Tests = legacyCountryEntry.PossibleTests
		}
		return res, nil
	}
	if jsonErr := json.Unmarshal(payload, &res); jsonErr != nil {
		return res, errors.Wrap(jsonErr, "error decoding json from DB")
	}
	return res, nil
}

// Parse parses country / state data from JSON
func Parse(payload []byte) (CollectionEntry, error) {
	res, err := parse(payload)
	if err != nil {
		return res, err
	}
	if res.When.Equal(time.Time{}) {
		return res, errors.Wrap(DateRequiredError, "Timestamp is required")
	}
	if res.When.Before(time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC)) {
		return res, errors.Wrap(InvalidDateError, "Timestamp is reported before the WHO official start date")
	}
	return res, nil
}

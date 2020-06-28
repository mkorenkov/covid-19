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
	Name          string    `json:"name"`
	When          time.Time `json:"when"`
	Cases         uint64    `json:"total_cases"`
	Deaths        uint64    `json:"total_deaths"`
	Tests         uint64    `json:"total_tests"`
	PossibleTests uint64    `json:"tests_per_1m"`
}

func parse(payload []byte) (DataEntry, error) {
	var res DataEntry
	legacyCountryEntry := legacyCountryData{}
	if legacyParseErr := json.Unmarshal(payload, &legacyCountryEntry); legacyParseErr == nil && legacyCountryEntry.PossibleTests > legacyCountryEntry.Tests {
		return DataEntry{
			Name:   legacyCountryEntry.Name,
			When:   legacyCountryEntry.When,
			Cases:  legacyCountryEntry.Cases,
			Deaths: legacyCountryEntry.Deaths,
			Tests:  legacyCountryEntry.PossibleTests,
		}, nil
	}
	if jsonErr := json.Unmarshal(payload, &res); jsonErr != nil {
		return res, errors.Wrap(jsonErr, "error decoding json from DB")
	}
	return res, nil
}

// Parse parses country / state data from JSON
func Parse(payload []byte) (DataEntry, error) {
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

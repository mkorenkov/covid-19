package documents

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/mkorenkov/covid-19/worldometers"
	"github.com/pkg/errors"
)

const (
	// StateCollection name of the states collection
	StateCollection = "States"
	// CountryCollection name of the countries collection
	CountryCollection = "Countries"
)

type CollectionEntry interface {
	GetWhen() time.Time
	GetName() string
	Save(w io.Writer) error
}

type DataEntry struct {
	Name   string    `json:"name"`
	When   time.Time `json:"when"`
	Cases  uint64    `json:"total_cases"`
	Deaths uint64    `json:"total_deaths"`
	Tests  uint64    `json:"total_tests"`
}

func (s DataEntry) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(s); err != nil {
		return errors.Wrapf(err, "error json encoding %s", s.GetName())
	}
	return nil
}

func (s DataEntry) GetWhen() time.Time {
	return s.When.UTC()
}

func (s DataEntry) GetName() string {
	return s.Name
}

func (s DataEntry) String() string {
	return fmt.Sprintf("%s %s[cases=%d tests=%d deaths=%d]", s.When.Format(time.RFC3339), s.Name, s.Cases, s.Tests, s.Deaths)
}

func FromState(state worldometers.UnitedState) *DataEntry {
	return &DataEntry{
		When:   time.Now(),
		Name:   state.Name,
		Cases:  state.TotalCases,
		Deaths: state.TotalDeaths,
		Tests:  state.TotalTests,
	}
}

func FromCountry(country worldometers.Country) *DataEntry {
	return &DataEntry{
		When:   time.Now(),
		Name:   country.Name,
		Cases:  country.TotalCases,
		Deaths: country.TotalDeaths,
		Tests:  country.TotalTests,
	}
}

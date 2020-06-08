package documents

import (
	"encoding/json"
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

type StateEntry struct {
	When time.Time `json:"when"`
	worldometers.UnitedState
}

func FromState(state worldometers.UnitedState) *StateEntry {
	return &StateEntry{
		time.Now(),
		state,
	}
}

func (s StateEntry) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(s); err != nil {
		return errors.Wrapf(err, "error json encoding %s", s.GetName())
	}
	return nil
}

func (s StateEntry) GetWhen() time.Time {
	return s.When
}

func (s StateEntry) GetName() string {
	return s.Name
}

type CountryEntry struct {
	When time.Time `json:"when"`
	worldometers.Country
}

func FromCountry(country worldometers.Country) *CountryEntry {
	return &CountryEntry{
		time.Now(),
		country,
	}
}

func (c CountryEntry) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(c); err != nil {
		return errors.Wrapf(err, "error json encoding %s", c.GetName())
	}
	return nil
}

func (c CountryEntry) GetWhen() time.Time {
	return c.When
}

func (c CountryEntry) GetName() string {
	return c.Name
}

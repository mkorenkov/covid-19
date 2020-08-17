package worldometers

import (
	"github.com/pkg/errors"
)

// UnitedState single row from worldometers
type UnitedState struct {
	Name        string `json:"name"`
	TotalCases  uint64 `json:"total_cases"`
	TotalDeaths uint64 `json:"total_deaths"`
	TotalTests  uint64 `json:"total_tests"`
}

func newStateFromRecord(data []string) (*UnitedState, error) {
	if data == nil || len(data) < 10 {
		return nil, errors.New("10 data items required to parse state")
	}

	totalCases, err := parseUint(data[2])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total cases")
	}
	totalDeaths, err := parseUint(data[4])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total deaths")
	}
	totalTests, err := parseUint(data[10])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total tests")
	}

	return &UnitedState{
		Name:        data[1],
		TotalCases:  totalCases,
		TotalDeaths: totalDeaths,
		TotalTests:  totalTests,
	}, nil
}

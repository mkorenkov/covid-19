package worldometers

import "github.com/pkg/errors"

// UnitedState single row from worldometers
type UnitedState struct {
	Name        string  `json:"name"`
	TotalCases  uint64  `json:"total_cases"`
	TotalDeaths uint64  `json:"total_deaths"`
	TotalTests  uint64  `json:"total_tests"`
	ActiveCases uint64  `json:"active_cases"`
	CasesPer1M  float64 `json:"cases_per_1m"`
	DeathsPer1M float64 `json:"deaths_per_1m"`
	TestsPer1M  float64 `json:"tests_per_1m"`
}

func newStateFromRecord(data []string) (*UnitedState, error) {
	if data == nil || len(data) < 10 {
		return nil, errors.New("10 data items required to parse state")
	}

	totalCases, err := parseUint(data[1])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total cases")
	}
	totalDeaths, err := parseUint(data[3])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total deaths")
	}
	totalTests, err := parseUint(data[8])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total tests")
	}
	var activeCases uint64
	// https://github.com/mkorenkov/covid-19/issues/1
	possibleNegativeActiveCases, err := parseInt(data[5])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse active cases")
	}
	if possibleNegativeActiveCases > 0 {
		activeCases, err = parseUint(data[5])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse active cases")
		}
	}
	cases1m, err := parseFloat(data[6])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse cases per 1M")
	}
	deaths1m, err := parseFloat(data[7])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse deaths per 1M")
	}
	tests1m, err := parseFloat(data[9])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse tests per 1M")
	}

	return &UnitedState{
		Name:        data[0],
		TotalCases:  totalCases,
		TotalDeaths: totalDeaths,
		TotalTests:  totalTests,
		ActiveCases: activeCases,
		CasesPer1M:  cases1m,
		DeathsPer1M: deaths1m,
		TestsPer1M:  tests1m,
	}, nil
}

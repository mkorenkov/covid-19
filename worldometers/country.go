package worldometers

import (
	"github.com/pkg/errors"
)

// Country single row from worldometers
type Country struct {
	Name           string  `json:"name"`
	TotalCases     uint64  `json:"total_cases"`
	TotalDeaths    uint64  `json:"total_deaths"`
	TotalRecovered uint64  `json:"total_recoverred"`
	TotalTests     uint64  `json:"total_tests"`
	ActiveCases    uint64  `json:"active_cases"`
	CriticalCases  uint64  `json:"critical_cases"`
	CasesPer1M     float64 `json:"cases_per_1m"`
	DeathsPer1M    float64 `json:"deaths_per_1m"`
	TestsPer1M     float64 `json:"tests_per_1m"`
	Population     uint64  `json:"population"`
	Region         string  `json:"region"`
}

func newCountryFromRecord(data []string) (*Country, error) {
	if data == nil || len(data) < 15 {
		return nil, errors.New("15 data items required to parse country")
	}

	totalCases, err := parseUint(data[2])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total cases")
	}
	totalDeaths, err := parseUint(data[4])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total deaths")
	}
	totalRecovered, err := parseUint(data[6])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total recoverred")
	}
	totalTests, err := parseFloat(data[12])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total tests")
	}
	var activeCases uint64
	// https://github.com/mkorenkov/covid-19/issues/1
	possibleNegativeActiveCases, err := parseInt(data[8])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse active cases")
	}
	if possibleNegativeActiveCases > 0 {
		activeCases, err = parseUint(data[8])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse active cases")
		}
	}
	var criticalCases uint64
	possibleNegativecriticalCases, err := parseInt(data[9])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse critical cases")
	}
	if possibleNegativecriticalCases > 0 {
		criticalCases, err = parseUint(data[9])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse critical cases")
		}
	}
	cases1m, err := parseFloat(data[13])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse cases per 1M")
	}
	deaths1m, err := parseFloat(data[11])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse deaths per 1M")
	}
	tests1m, err := parseFloat(data[10])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse tests per 1M")
	}
	population, err := parseUint(data[14])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse population")
	}

	return &Country{
		Name:           data[1],
		TotalCases:     totalCases,
		TotalDeaths:    totalDeaths,
		TotalRecovered: totalRecovered,
		TotalTests:     uint64(totalTests),
		ActiveCases:    activeCases,
		CriticalCases:  criticalCases,
		CasesPer1M:     cases1m,
		DeathsPer1M:    deaths1m,
		TestsPer1M:     tests1m,
		Population:     population,
		Region:         data[15],
	}, nil
}

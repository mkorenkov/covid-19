package worldofmeters

import (
	"github.com/pkg/errors"
)

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
	Region         string  `json:"region"`
}

func NewCountryFromRecord(data []string) (*Country, error) {
	if data == nil || len(data) < 10 {
		return nil, errors.New("13 data items required to parse country")
	}

	totalCases, err := parseUint(data[1])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total cases")
	}
	totalDeaths, err := parseUint(data[3])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total deaths")
	}
	totalRecovered, err := parseUint(data[5])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total recoverred")
	}
	totalTests, err := parseUint(data[10])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total tests")
	}
	activeCases, err := parseUint(data[6])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse active cases")
	}
	criticalCases, err := parseUint(data[7])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse critical cases")
	}
	cases1m, err := parseFloat(data[8])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse cases per 1M")
	}
	deaths1m, err := parseFloat(data[9])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse deaths per 1M")
	}
	tests1m, err := parseFloat(data[11])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse tests per 1M")
	}

	return &Country{
		Name:           data[0],
		TotalCases:     totalCases,
		TotalDeaths:    totalDeaths,
		TotalRecovered: totalRecovered,
		TotalTests:     totalTests,
		ActiveCases:    activeCases,
		CriticalCases:  criticalCases,
		CasesPer1M:     cases1m,
		DeathsPer1M:    deaths1m,
		TestsPer1M:     tests1m,
		Region:         data[12],
	}, nil
}

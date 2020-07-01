package worldometers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUSA(t *testing.T) {
	input := `1;USA;2,007,449;;112,469;;761,708;;1,133,272;16,923;6,067;340;21,291,677;64,349;330,880,530;North America;165;2,942;16`
	res, err := newCountryFromRecord(strings.Split(input, ";"))
	require.Nil(t, err)
	assert.Equal(t, "USA", res.Name)
	assert.Equal(t, uint64(2007449), res.TotalCases, "total cases")
	assert.Equal(t, uint64(112469), res.TotalDeaths, "total deaths")
	assert.Equal(t, uint64(761708), res.TotalRecovered, "total recovered")
	assert.Equal(t, uint64(1133272), res.ActiveCases, "active cases")
	assert.Equal(t, uint64(16923), res.CriticalCases, "critical cases")
	assert.Equal(t, float64(6067), res.CasesPer1M, "cases per 1M")
	assert.Equal(t, float64(340), res.DeathsPer1M, "deaths per 1M")
	assert.Equal(t, uint64(21291677), res.TotalTests, "total tests")
	assert.Equal(t, float64(64349), res.TestsPer1M, "tests per 1M")
	assert.Equal(t, uint64(330880530), res.Population, "population")
	assert.Equal(t, "North America", res.Region, "region")
}

func TestUkraine(t *testing.T) {
	input := `33;Ukraine;44,998;+664;1,173;+14;19,548;+433;24,277;97;1,029;27;666,147;15,232;43,732,279;Europe;972;37,282;66`
	res, err := newCountryFromRecord(strings.Split(input, ";"))
	require.Nil(t, err)
	assert.Equal(t, "Ukraine", res.Name)
	assert.Equal(t, 44998, int(res.TotalCases), "total cases")
	assert.Equal(t, 1173, int(res.TotalDeaths), "total deaths")
	assert.Equal(t, 19548, int(res.TotalRecovered), "total recovered")
	assert.Equal(t, 24277, int(res.ActiveCases), "active cases")
	assert.Equal(t, 97, int(res.CriticalCases), "critical cases")
	assert.Equal(t, 1029, int(res.CasesPer1M), "cases per 1M")
	assert.Equal(t, 27, int(res.DeathsPer1M), "deaths per 1M")
	assert.Equal(t, 666147, int(res.TotalTests), "total tests")
	assert.Equal(t, 15232, int(res.TestsPer1M), "tests per 1M")
	assert.Equal(t, 43732279, int(res.Population), "population")
	assert.Equal(t, "Europe", res.Region, "region")
}

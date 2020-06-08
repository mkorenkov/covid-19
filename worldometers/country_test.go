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
	assert.Equal(t, float64(64349), res.CasesPer1M, "cases per 1M")
	assert.Equal(t, float64(340), res.DeathsPer1M, "deaths per 1M")
	assert.Equal(t, uint64(21291677), res.TotalTests, "total tests")
	assert.Equal(t, float64(6067), res.TestsPer1M, "tests per 1M")
	assert.Equal(t, uint64(330880530), res.Population, "population")
	assert.Equal(t, "North America", res.Region, "region")
}

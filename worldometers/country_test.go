package worldometers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUSA(t *testing.T) {
	input := []string{
		"USA",
		"532,879",
		"",
		"20,577",
		"",
		"30,453",
		"481,849",
		"11,471",
		"1,610",
		"62",
		"2,670,674",
		"8,068",
		"North America",
	}
	res, err := newCountryFromRecord(input)
	require.Nil(t, err)
	assert.Equal(t, "USA", res.Name)
	assert.Equal(t, uint64(532879), res.TotalCases)
	assert.Equal(t, uint64(20577), res.TotalDeaths)
	assert.Equal(t, uint64(30453), res.TotalRecovered)
	assert.Equal(t, uint64(481849), res.ActiveCases)
	assert.Equal(t, uint64(11471), res.CriticalCases)
	assert.Equal(t, float64(1610), res.CasesPer1M)
	assert.Equal(t, float64(62), res.DeathsPer1M)
	assert.Equal(t, uint64(2670674), res.TotalTests)
	assert.Equal(t, float64(8068), res.TestsPer1M)
	assert.Equal(t, "North America", res.Region)
}

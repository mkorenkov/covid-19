package worldometers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUSA(t *testing.T) {
	input := []string{
		"1",
		"USA",
		"1,484,285",
		"",
		"88,507",
		"",
		"327,751",
		"1,068,027",
		"16,139",
		"4,488",
		"268",
		"11,090,900",
		"33,532",
		"330,758,784",
		"North America",
	}
	res, err := newCountryFromRecord(input)
	require.Nil(t, err)
	assert.Equal(t, "USA", res.Name)
	assert.Equal(t, uint64(1484285), res.TotalCases)
	assert.Equal(t, uint64(88507), res.TotalDeaths)
	assert.Equal(t, uint64(327751), res.TotalRecovered)
	assert.Equal(t, uint64(1068027), res.ActiveCases)
	assert.Equal(t, uint64(16139), res.CriticalCases)
	assert.Equal(t, float64(4488), res.CasesPer1M)
	assert.Equal(t, float64(268), res.DeathsPer1M)
	assert.Equal(t, uint64(11090900), res.TotalTests)
	assert.Equal(t, float64(33532), res.TestsPer1M)
	assert.Equal(t, uint64(330758784), res.Population)
	assert.Equal(t, "North America", res.Region)
}

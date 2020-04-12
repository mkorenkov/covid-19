package worldometers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNY(t *testing.T) {
	input := []string{
		"New York",
		"181,144",
		"",
		"8,627",
		"",
		"155,840",
		"9,233",
		"440",
		"440,980",
		"22,478",
	}
	res, err := newStateFromRecord(input)
	require.Nil(t, err)
	assert.Equal(t, "New York", res.Name)
	assert.Equal(t, uint64(181144), res.TotalCases)
	assert.Equal(t, uint64(8627), res.TotalDeaths)
	assert.Equal(t, uint64(155840), res.ActiveCases)
	assert.Equal(t, float64(9233), res.CasesPer1M)
	assert.Equal(t, float64(440), res.DeathsPer1M)
	assert.Equal(t, uint64(440980), res.TotalTests)
	assert.Equal(t, float64(22478), res.TestsPer1M)
}

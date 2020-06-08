package worldometers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNY(t *testing.T) {
	input := `New York;398,828;;30,442;;283,695;20,502;1,565;2,497,842;128,400;[;[`
	res, err := newStateFromRecord(strings.Split(input, ";"))
	require.Nil(t, err)
	assert.Equal(t, "New York", res.Name)
	assert.Equal(t, uint64(398828), res.TotalCases)
	assert.Equal(t, uint64(30442), res.TotalDeaths)
	assert.Equal(t, uint64(283695), res.ActiveCases)
	assert.Equal(t, float64(20502), res.CasesPer1M)
	assert.Equal(t, float64(1565), res.DeathsPer1M)
	assert.Equal(t, uint64(2497842), res.TotalTests)
	assert.Equal(t, float64(128400), res.TestsPer1M)
}

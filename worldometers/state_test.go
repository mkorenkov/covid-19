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

func TestCalifornia(t *testing.T) {
	input := `California;230,891;;6,081;;161,148;5,844;154;4,168,509;105,499;[;[`
	res, err := newStateFromRecord(strings.Split(input, ";"))
	require.Nil(t, err)
	assert.Equal(t, "California", res.Name)
	assert.Equal(t, 230891, int(res.TotalCases))
	assert.Equal(t, 6081, int(res.TotalDeaths))
	assert.Equal(t, 161148, int(res.ActiveCases))
	assert.Equal(t, 5844, int(res.CasesPer1M))
	assert.Equal(t, 154, int(res.DeathsPer1M))
	assert.Equal(t, 4168509, int(res.TotalTests))
	assert.Equal(t, 105499, int(res.TestsPer1M))
}

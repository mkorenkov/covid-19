package worldometers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAug16ChangesUSATotal(t *testing.T) {
	input := `;USA Total;5,565,461;+35,672;173,096;+490;2,921,070;2,471,295;16,814;523;70,942,037;214,325;;;`
	res, err := newStateFromRecord(strings.Split(input, ";"))
	require.Nil(t, err)
	assert.Equal(t, "USA Total", res.Name)
	assert.Equal(t, 5565461, int(res.TotalCases))
	assert.Equal(t, 173096, int(res.TotalDeaths))
	assert.Equal(t, 70942037, int(res.TotalTests))
}

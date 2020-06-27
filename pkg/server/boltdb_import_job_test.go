package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLegacyCountry(t *testing.T) {
	data := `{"name":"USA","total_cases":2026493,"total_deaths":113055,"total_recoverred":773480,"total_tests":342,"active_cases":0,"critical_cases":1139958,"cases_per_1m":16907,"deaths_per_1m":6124,"tests_per_1m":21725064,"population":65657,"region":"330,885,824"}`
	res, err := parseCountry([]byte(data))
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(2026493), res.Cases)
	assert.Equal(t, uint64(113055), res.Deaths)
	assert.Equal(t, uint64(21725064), res.Tests)
}

func TestParseCountry(t *testing.T) {
	data := `
		{
			"name": "USA",
			"when": "2020-06-26T23:23:35.383642-07:00",
			"total_cases": 2552956,
			"total_deaths": 127640,
			"total_tests": 31352500
		}
	`
	res, err := parseCountry([]byte(data))
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(2552956), res.Cases)
	assert.Equal(t, uint64(127640), res.Deaths)
	assert.Equal(t, uint64(31352500), res.Tests)
}

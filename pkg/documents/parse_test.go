package documents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	data   string
	name   string
	cases  int
	deaths int
	tests  int
}

func TestParseCountriesAndStates(t *testing.T) {
	testCases := []testData{
		{name: "USA", cases: 2026493, deaths: 113055, tests: 21725064, data: `{"name":"USA", "when": "2020-06-01T22:47:18.753632159Z", "total_cases":2026493,"total_deaths":113055,"total_recoverred":773480,"total_tests":342,"active_cases":0,"critical_cases":1139958,"cases_per_1m":16907,"deaths_per_1m":6124,"tests_per_1m":21725064,"population":65657,"region":"330,885,824"}`},
		{name: "USA", cases: 2552956, deaths: 127640, tests: 31352500, data: `{"name": "USA", "when": "2020-06-26T23:23:35.383642-07:00", "total_cases": 2552956, "total_deaths": 127640, "total_tests": 31352500}`},
		{name: "California", cases: 22409, deaths: 633, tests: 182986, data: `{"name":"California", "when": "2020-06-01T08:07:21.288909973Z", "total_cases":22409,"total_deaths":633,"total_tests":182986,"active_cases":20836,"cases_per_1m":572,"deaths_per_1m":16,"tests_per_1m":4674}`},
		{name: "New York", cases: 381887, deaths: 30078, tests: 2167831, data: `{"name":"New York", "when": "2020-06-01T08:07:21.288909973Z", "total_cases":381887,"total_deaths":30078,"total_tests":2167831,"active_cases":284559,"cases_per_1m":19631,"deaths_per_1m":1546,"tests_per_1m":111436}`},
		{name: "Florida", cases: 132545, deaths: 3392, tests: 1830791, data: `{"name": "Florida", "when": "2020-06-28T01:08:48.813843-07:00", "total_cases": 132545, "total_deaths": 3392, "total_tests": 1830791}`},
	}
	for _, testCase := range testCases {
		res, err := Parse([]byte(testCase.data))
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.True(t, res.When.After(time.Date(2020, 01, 01, 01, 01, 01, 01, time.Now().Location())))
		assert.Equal(t, testCase.name, res.Name)
		assert.Equal(t, testCase.cases, int(res.Cases))
		assert.Equal(t, testCase.deaths, int(res.Deaths))
		assert.Equal(t, testCase.tests, int(res.Tests))
	}
}

func TestParseInvalidDate(t *testing.T) {
	testCases := []string{
		`{"name": "USA", "total_cases":2026493,"total_deaths":113055,"total_recoverred":773480,"total_tests":342,"active_cases":0,"critical_cases":1139958,"cases_per_1m":16907,"deaths_per_1m":6124,"tests_per_1m":21725064,"population":65657,"region":"330,885,824"}`,
		`{"name": "USA", "when": "2018-06-01T22:47:18.753632159Z", "total_cases": 2552956, "total_deaths": 127640, "total_tests": 31352500}`,
	}
	for _, testCase := range testCases {
		_, err := Parse([]byte(testCase))
		require.Error(t, err)
	}
}

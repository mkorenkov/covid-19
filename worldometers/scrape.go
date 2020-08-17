package worldometers

import (
	"context"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

const (
	countriesURL = "https://www.worldometers.info/coronavirus/"
	statesURL    = "https://www.worldometers.info/coronavirus/country/us/"
)

// HTTPClient common interface for many HTTP clients, including http.client from stdlib.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// htmlTableToArrays converts given table to text only array of arrays
func htmlTableToArrays(rows []*html.Node) [][]string {
	tableData := make([][]string, len(rows))

	for idx, tr := range rows {
		rowData := []string{}
		for td := tr.FirstChild; td != nil; td = td.NextSibling {
			if td.Data == "th" || td.Data == "td" {
				rowData = append(rowData, readText(td))
			}
		}
		tableData[idx] = rowData
	}

	return tableData
}

// readText traverses HTML tree and reads the inner text
func readText(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := readText(c)
		if result != "" {
			return result
		}
	}
	return ""
}

// Countries scrapes worldometers and returns per country information.
func Countries(ctx context.Context, httpclient HTTPClient) (map[string]*Country, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", countriesURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating HTTP request")
	}
	req = req.WithContext(ctx)

	res, err := httpclient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failure")
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "goquery error")
	}

	rows := []*html.Node{}
	doc.Find("#main_table_countries_today").Find("tbody").Find("tr").Each(func(i int, trSel *goquery.Selection) {
		rows = append(rows, trSel.Nodes...)
	})
	srcTable := htmlTableToArrays(rows)
	dataSource := map[string]*Country{}
	for _, row := range srcTable {
		if len(row) > 1 {
			record, err := newCountryFromRecord(row)
			if err != nil {
				return nil, errors.Wrapf(err, "country parse error, data row: '%v'", strings.Join(row, ";"))
			}
			dataSource[record.Name] = record
		}
	}
	return dataSource, nil
}

// States scrapes worldometers and returns per state information.
func States(ctx context.Context, httpclient HTTPClient) (map[string]*UnitedState, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", statesURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating HTTP request")
	}
	req = req.WithContext(ctx)

	res, err := httpclient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failure")
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "goquery error")
	}

	rows := []*html.Node{}
	doc.Find("#usa_table_countries_today").Find("tbody").Find("tr").Each(func(i int, trSel *goquery.Selection) {
		rows = append(rows, trSel.Nodes...)
	})
	srcTable := htmlTableToArrays(rows)
	dataSource := map[string]*UnitedState{}
	for _, row := range srcTable {
		if len(row) > 1 {
			record, err := newStateFromRecord(row)
			if err != nil {
				return nil, errors.Wrapf(err, "state parse error, data row: '%v'", strings.Join(row, ";"))
			}
			dataSource[record.Name] = record
		}
	}
	return dataSource, nil
}

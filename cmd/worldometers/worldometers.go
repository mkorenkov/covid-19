package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mkorenkov/covid-19-parser/worldofmeters"
	"golang.org/x/net/html"
)

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

func scrapeCountries() {
	res, err := http.Get("https://www.worldometers.info/coronavirus/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	rows := []*html.Node{}
	doc.Find("#main_table_countries_today").Find("tbody").Find("tr").Each(func(i int, trSel *goquery.Selection) {
		rows = append(rows, trSel.Nodes...)
	})
	srcTable := htmlTableToArrays(rows)
	dataSource := map[string]interface{}{}
	for _, row := range srcTable {
		if len(row) > 1 {
			record, err := worldofmeters.NewCountryFromRecord(row)
			if err != nil {
				log.Fatal(err)
			}
			dataSource[row[0]] = record
		}
	}
	fmt.Println(dataSource["USA"])
	fmt.Println(dataSource["Russia"])
}

func scrapeUSA() {
	res, err := http.Get("https://www.worldometers.info/coronavirus/country/us/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	rows := []*html.Node{}
	doc.Find("#usa_table_countries_today").Find("tbody").Find("tr").Each(func(i int, trSel *goquery.Selection) {
		rows = append(rows, trSel.Nodes...)
	})
	srcTable := htmlTableToArrays(rows)
	dataSource := map[string]interface{}{}
	for _, row := range srcTable {
		if len(row) > 1 {
			record, err := worldofmeters.NewStateFromRecord(row)
			if err != nil {
				log.Fatal(err)
			}
			dataSource[row[0]] = record
		}
	}
	fmt.Println(dataSource["New York"])
	fmt.Println(dataSource["California"])
}

func main() {
	scrapeCountries()
}

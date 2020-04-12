package main

import (
	"log"

	"github.com/mkorenkov/covid-19/worldometers"
)

func main() {
	countries, err := worldometers.Countries()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(countries["USA"])
}

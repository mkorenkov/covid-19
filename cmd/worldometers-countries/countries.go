package main

import (
	"context"
	"log"

	"github.com/mkorenkov/covid-19/worldometers"
)

func main() {
	countries, err := worldometers.Countries(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println(countries["USA"])
}

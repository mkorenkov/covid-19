package main

import (
	"context"
	"log"

	"github.com/mkorenkov/covid-19/worldometers"
)

func main() {
	states, err := worldometers.States(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println(states["California"])
}

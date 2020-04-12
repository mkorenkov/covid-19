package main

import (
	"log"

	"github.com/mkorenkov/covid-19/worldometers"
)

func main() {
	states, err := worldometers.States()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(states["California"])
}

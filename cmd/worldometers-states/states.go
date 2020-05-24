package main

import (
	"context"
	"log"
	"net/http"

	"github.com/mkorenkov/covid-19/worldometers"
)

func main() {
	states, err := worldometers.States(context.Background(), http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(states["California"])
}

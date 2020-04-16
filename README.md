# covid-19

wordometers COVID-19 data HTML scraper.

## Countries data
```
countries, err := worldometers.Countries(context.Background())
if err != nil {
    log.Fatal(err)
}
log.Println(countries["USA"])
```

## State data
```
states, err := worldometers.States(context.Background())
if err != nil {
    log.Fatal(err)
}
log.Println(states["California"])
```

## Examples
See `cmd/` directory.

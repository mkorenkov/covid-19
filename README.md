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

## Daemon mode

```
#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# S3 compatible storage
export COVIDDY_S3_ACCESS_KEY="<access_key>"
export COVIDDY_S3_SECRET="<secret>"
export COVIDDY_S3_REGION="us-west-1"
export COVIDDY_S3_ENDPOINT="https://s3.us-west-1.wasabisys.com"
export COVIDDY_S3_BUCKET="covid-19[/dev]"
# /api/v1/internal HTTP Basic Auth settings
export COVIDDY_CREDENTIALS="user1:password1,user2:password2"
# location for the boltdb storage dir
export COVIDDY_STORAGE_DIR="/tmp/data/covid-19"

go run cmd/worldometersd/main.go
```

FROM golang:1.14 AS build
WORKDIR /go/src/github.com/mkorenkov/covid-19
COPY . .
RUN cd /go/src/github.com/mkorenkov/covid-19 && CGO_ENABLED=0 GO111MODULE=off GOOS=linux go build -o bin/coviddy cmd/coviddy/main.go

FROM alpine:latest
ENV COVIDDY_STORAGE_DIR="/srv/coviddy"
ENV COVIDDY_LISTEN_ADDR=":9898"
ENV COVIDDY_SCRAPE_INTERVAL="163m"
# explicitly set user/group IDs
RUN addgroup -S -g 998 coviddy && \
    adduser -S -h /srv/coviddy -u 998 -G coviddy coviddy && \
    apk add --update \
        bash \
        bind-tools \
        ca-certificates \
        su-exec \
        tzdata
COPY --from=build /go/src/github.com/mkorenkov/covid-19/bin/coviddy /bin/coviddy
VOLUME ["/srv/coviddy"]
ENTRYPOINT ["su-exec", "coviddy"]
CMD ["/bin/coviddy"]

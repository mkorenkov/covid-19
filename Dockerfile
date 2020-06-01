FROM golang:1.14 AS build
WORKDIR /go/src/github.com/mkorenkov/covid-19
COPY . .
RUN cd /go/src/github.com/mkorenkov/covid-19 && CGO_ENABLED=0 GO111MODULE=off GOOS=linux go build -o bin/coviddy cmd/coviddy/main.go

FROM alpine:latest
ENV STORAGE_DIR="/srv/coviddy"
# explicitly set user/group IDs
RUN addgroup -S -g 993 coviddy && \
    adduser -S -h /srv/coviddy -u 993 -G coviddy coviddy && \
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

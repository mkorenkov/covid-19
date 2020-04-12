FROM golang:1.14 AS build
WORKDIR /go/src/github.com/mkorenkov/covid-19
COPY . .
RUN cd /go/src/github.com/mkorenkov/covid-19 && CGO_ENABLED=0 GO111MODULE=off GOOS=linux go build -o bin/as_json cmd/worldometers-save-json/as_json.go

FROM alpine:latest AS certs
RUN apk --update add ca-certificates

FROM alpine:latest
ENV STORAGE_DIR="/srv/data/covid-19"
RUN apk add bind-tools
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/github.com/mkorenkov/covid-19/bin/as_json /bin/as_json
CMD ["/bin/as_json"]

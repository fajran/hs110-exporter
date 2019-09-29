FROM golang:1.12 AS build

WORKDIR /go/src/github.com/fajran/hs110-exporter

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go install github.com/fajran/hs110-exporter

FROM scratch
COPY --from=build /go/bin/hs110-exporter /hs110-exporter
ENTRYPOINT ["/hs110-exporter"]


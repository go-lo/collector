[![Go Report Card](https://goreportcard.com/badge/github.com/go-lo/collector)](https://goreportcard.com/report/github.com/go-lo/collector)
[![Build Status](https://travis-ci.com/go-lo/collector.svg?branch=master)](https://travis-ci.com/go-lo/collector)
[![GoDoc](https://godoc.org/github.com/go-lo/collector?status.svg)](https://godoc.org/github.com/go-lo/collector)

# Collector

Collector provides an API which:

 * Takes a [golo.Output](https://godoc.org/github.com/go-lo/go-lo#Output) in json
 * Takes a database name from the request path
 * Pushes this information into a channel

This channel is, in a gofunc;

 * Consumed
 * Pushed to various storage locations

Currently we support [influx](https://www.influxdata.com/), but adding new storage locations is very simple; the interface is:

```go
type Collector interface {
    Push(OutputMapper) error
}
```

## Usage

This project is best used with docker:

```bash
$ docker run goload/collector --help
Usage of /collector:
  -influx string
        influx host (default "http://localhost:8086")
```

The default options will point to influx on localhost, the flag `-influx` will change this.


## Development

This project strives for high test coverage, and for happy and sad paths to be covered. Please do ensure pull requests have tests with them, where appropriate.

Building and testing can be done with either the standard docker toolchain:

```bash
$ go get -u
$ go test -v
$ go build
```

Or via the convenience wrappers in the `Makefile`

```bash
$ make test collector
```

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/jspc/loadtest"
)

type OutputWriteWrapper struct {
	output   loadtest.Output
	database string
}

var (
	influx = flag.String("influx", "http://localhost:8086", "influx host")
)

func main() {
	flag.Parse()

	idb, err := NewInfluxdbCollector(*influx, "magnum")
	if err != nil {
		panic(err)
	}

	collectors := []Collector{
		idb,
	}

	c := make(chan OutputWriteWrapper)

	go func() {
		for o := range c {
			for _, collector := range collectors {
				err = collector.Push(o)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}()

	api := API{
		OutputChan: c,
	}

	panic(http.ListenAndServe(":8082", api))
}

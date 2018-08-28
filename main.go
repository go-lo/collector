package main

import (
	"flag"
	"log"

	"github.com/go-lo/go-lo"
	"github.com/valyala/fasthttp"
)

// OutputMapper includes a golo.Output and additional
// collector metadata
type OutputMapper struct {
	output   golo.Output
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

	c := make(chan OutputMapper)

	go func() {
		for o := range c {
			for _, collector := range collectors {
				go func() {
					err = collector.Push(o)
					if err != nil {
						log.Print(err)
					}
				}()
			}
		}
	}()

	api := API{
		OutputChan: c,
	}

	panic(fasthttp.ListenAndServe(":8082", api.Route))
}

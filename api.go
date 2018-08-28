package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-lo/go-lo"
	"github.com/valyala/fasthttp"
)

// API provides an http interface into the collector;
// specifically pushing metrics through
type API struct {
	OutputChan chan OutputMapper
}

func (a API) Route(ctx *fasthttp.RequestCtx) {
	switch {
	case strings.HasPrefix(string(ctx.Path()), "/push/"):
		a.Push(ctx)

	default:
		ctx.Error(fmt.Sprintf("%s not found", string(ctx.Path())), fasthttp.StatusNotFound)
	}
}

// Push handles 'POST example.com/push/$DB' requests- it receives
// them, determines the database (from the path) and then pushes
// into a channel for different collectors to consume and push
func (a API) Push(ctx *fasthttp.RequestCtx) {
	index := strings.TrimPrefix(string(ctx.Path()), "/push/")
	if index == "" {
		ctx.Error(fmt.Sprintf("%s not found", string(ctx.Path())), fasthttp.StatusNotFound)

		return
	}

	if string(ctx.Method()) != http.MethodPost {
		ctx.Error("method not allowed", fasthttp.StatusMethodNotAllowed)

		return
	}

	body := ctx.Request.Body()
	if len(body) == 0 {
		ctx.Error("no body found", fasthttp.StatusBadRequest)

		return
	}

	ol := new([]golo.Output)

	err := json.Unmarshal(body, ol)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)

		return
	}

	for _, o := range *ol {
		a.OutputChan <- OutputMapper{
			output:   o,
			database: index,
		}
	}
}

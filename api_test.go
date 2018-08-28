package main

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestServeHTTP(t *testing.T) {
	for _, test := range []struct {
		name         string
		method       string
		path         string
		body         string
		destroyBody  bool
		expectStatus int
	}{
		{"valid request", "POST", "/push/blah", `[{}]`, false, 200},
		{"path not found", "POST", "/hello", ``, false, 404},
		{"invalid method", "PATCH", "/push/blah", ``, false, 405},
		{"missing database", "POST", "/push/", `[{}]`, false, 404},
		{"bad json", "POST", "/push/blah", "}}", false, 400},
		{"malformed request body", "POST", "/push/blah", "", true, 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			a := API{make(chan OutputMapper)}
			defer close(a.OutputChan)

			go func() {
				for {
					<-a.OutputChan
				}
			}()

			u, _ := url.Parse("http://example.com")
			u.Path = test.path

			r := new(fasthttp.Request)
			r.Header.SetMethod(test.method)
			r.SetRequestURI(u.String())
			r.SetBodyStream(bytes.NewBufferString(test.body), len([]byte(test.body)))

			ctx := &fasthttp.RequestCtx{Request: *r}

			a.Route(ctx)

			status := ctx.Response.Header.StatusCode()
			if test.expectStatus != status {
				t.Errorf("expected %d, received %d", test.expectStatus, status)
			}
		})
	}
}

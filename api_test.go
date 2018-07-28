package main

import (
	"bytes"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	for _, test := range []struct {
		name         string
		method       string
		path         string
		body         string
		expectStatus int
	}{
		{"valid request", "POST", "/push/blah", `{}`, 200},
		{"path not found", "POST", "/hello", ``, 404},
		{"invalid method", "PATCH", "/push/blah", ``, 405},
		{"missing database", "POST", "/push/", `{}`, 404},
		{"bad json", "POST", "/push/blah", "}}", 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			a := API{make(chan OutputWriteWrapper)}
			defer close(a.OutputChan)

			go func() {
				for {
					<-a.OutputChan
				}
			}()

			u, _ := url.Parse("http://example.com")
			u.Path = test.path

			r := httptest.NewRequest(test.method, u.String(), bytes.NewBufferString(test.body))
			w := httptest.NewRecorder()

			a.ServeHTTP(w, r)

			if test.expectStatus != w.Code {
				t.Errorf("expected %d, received %d", test.expectStatus, w.Code)
			}
		})
	}
}

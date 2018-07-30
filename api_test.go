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

			r := httptest.NewRequest(test.method, u.String(), bytes.NewBufferString(test.body))
			w := httptest.NewRecorder()

			if test.destroyBody {
				r.Body = nil
			}

			a.ServeHTTP(w, r)

			if test.expectStatus != w.Code {
				t.Errorf("expected %d, received %d", test.expectStatus, w.Code)
			}
		})
	}
}

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/jspc/loadtest"
)

type dummyInfluxClient struct {
	lastBody []byte
	status   int
	err      bool
	dropReq  bool
}

func (d *dummyInfluxClient) PostForm(u string, i url.Values) (resp *http.Response, err error) {
	d.lastBody = []byte(i.Get("q"))

	return d.ret(u)
}

func (d *dummyInfluxClient) Post(u string, _ string, b io.Reader) (resp *http.Response, err error) {
	d.lastBody, _ = ioutil.ReadAll(b)

	return d.ret(u)
}

func (d *dummyInfluxClient) ret(u string) (resp *http.Response, err error) {
	if d.status == 0 {
		d.status = 200
	}

	resp = &http.Response{StatusCode: d.status, Body: ioutil.NopCloser(bytes.NewBufferString("some message"))}

	if !d.dropReq {
		reqURL, _ := url.Parse(u)

		resp.Request = &http.Request{Method: "POST", URL: reqURL}
	}

	if d.err {
		err = fmt.Errorf("an error")
		resp = nil
	}

	return
}

func TestNewInfluxCollector(t *testing.T) {
	_, err := NewInfluxdbCollector("example.com", "test")

	if err != nil {
		t.Errorf("unexpected error %+v", err)
	}
}

func TestInflux_CreateIndex(t *testing.T) {
	for _, test := range []struct {
		name        string
		db          string
		client      httpClient
		expect      string
		expectError bool
	}{
		{"valid request", "a-db", &dummyInfluxClient{}, "CREATE DATABASE a-db", false},
		{"network error", "a-db", &dummyInfluxClient{err: true}, "CREATE DATABASE a-db", true},
		{"bad req", "a-db", &dummyInfluxClient{status: 500}, "CREATE DATABASE a-db", true},

		// I don't remember the circumstances for writing a check for whether the
		// request in a response was nil, but I do remember it was a pretty irritating
		// bug to find.
		//
		// Thus: there's a condition in the influx collector so we should test it
		{"some weird thing", "a-db", &dummyInfluxClient{status: 500, dropReq: true}, "CREATE DATABASE a-db", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			i, _ := NewInfluxdbCollector("example.com", "test")
			i.client = test.client

			err := i.CreateIndex(test.db)
			if test.expectError && err == nil {
				t.Errorf("expected error")
			}

			if !test.expectError && err != nil {
				t.Errorf("unexpected error %+v", err)
			}

			lb := string(i.client.(*dummyInfluxClient).lastBody)
			if test.expect != lb {
				t.Errorf("expected %q received %q", test.expect, lb)
			}
		})
	}
}

func TestInflux_Push(t *testing.T) {
	n := time.Now()

	o := loadtest.Output{
		URL:       "example.com",
		Method:    "DELETE",
		Status:    http.StatusTeapot,
		Error:     nil,
		Size:      420 * 69,
		Duration:  1000000,
		Timestamp: n,
	}
	oO := fmt.Sprintf("request,url=example.com,method=DELETE,status=418,error=false size=28980,duration=1000000 %d", n.UnixNano())

	for _, test := range []struct {
		name        string
		ow          OutputMapper
		client      httpClient
		indices     map[string]byte
		expect      string
		expectError bool
	}{
		{"well formed output, first push", OutputMapper{o, "a-db"}, &dummyInfluxClient{}, make(map[string]byte), oO, false},
		{"well formed output, not first", OutputMapper{o, "a-db"}, &dummyInfluxClient{}, map[string]byte{"a-db": '1'}, oO, false},
		{"bad response", OutputMapper{o, "a-db"}, &dummyInfluxClient{status: 500}, map[string]byte{"a-db": '1'}, oO, true},
		{"network error", OutputMapper{o, "a-db"}, &dummyInfluxClient{err: true}, map[string]byte{"a-db": '1'}, oO, true},
		{"missing/ unfinished", OutputMapper{loadtest.Output{}, "a-db"}, &dummyInfluxClient{}, map[string]byte{"a-db": '1'}, "", true},

		// See above for explanation, such that it is
		{"weirdness", OutputMapper{o, "a-db"}, &dummyInfluxClient{status: 500, dropReq: true}, map[string]byte{"a-db": '1'}, oO, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			i, _ := NewInfluxdbCollector("example.com", "test")
			i.client = test.client
			i.indices = test.indices

			err := i.Push(test.ow)
			if test.expectError && err == nil {
				t.Errorf("expected error")
			}

			if !test.expectError && err != nil {
				t.Errorf("unexpected error %+v", err)
			}

			lb := string(i.client.(*dummyInfluxClient).lastBody)
			if test.expect != lb {
				t.Errorf("expected %q received %q", test.expect, lb)
			}

		})
	}
}

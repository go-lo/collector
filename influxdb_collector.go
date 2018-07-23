package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/jspc/loadtest"
)

type httpClient interface {
	Post(string, string, io.Reader) (*http.Response, error)
}

type InfluxdbCollector struct {
	Host     string
	Database string

	client httpClient
}

func NewInfluxdbCollector(host, db string) (c InfluxdbCollector, err error) {
	c = InfluxdbCollector{
		Host:     host,
		Database: db,
		client:   new(http.Client),
	}

	err = c.SetIndex()

	return
}

func (c InfluxdbCollector) SetIndex() (err error) {
	t := "q=CREATE DATABASE {{.Database}}"

	q, err := c.tmpl(t)
	if err != nil {
		return
	}

	return c.post("query", url.PathEscape(q))
}

func (c InfluxdbCollector) Push(o loadtest.Output) (err error) {
	t := "request,url={{.URL}},method={{.Method}},status={{.Status}},error={{if .Error}}true{{else}}false{{end}} size={{.Size}},duration={{.Size}} {{.Timestamp}}"
	q, err := c.tmpl(t)
	if err != nil {
		return
	}

	p := fmt.Sprintf("write?db=%s", c.Database)

	return c.post(p, q)
}

func (c InfluxdbCollector) tmpl(s string) (out string, err error) {
	tmpl, err := template.New("create").Parse(s)
	if err != nil {
		return
	}

	outputBuffer := bytes.NewBuffer(make([]byte, 0))

	err = tmpl.Execute(outputBuffer, c)
	if err != nil {
		return
	}

	out = outputBuffer.String()

	return
}

func (c InfluxdbCollector) post(path, query string) (err error) {
	fmt.Println(query)

	r := strings.NewReader(query)

	resp, err := c.client.Post(fmt.Sprintf("%s/%s", c.Host, path), "text/plain", r)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	fmt.Println(string(body))

	return
}

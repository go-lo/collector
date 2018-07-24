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
	PostForm(string, url.Values) (*http.Response, error)
}

type InfluxdbCollector struct {
	Host     string
	Database string

	client httpClient
	funcs  template.FuncMap
}

func NewInfluxdbCollector(host, db string) (c InfluxdbCollector, err error) {
	c = InfluxdbCollector{
		Host:     host,
		Database: db,
		client:   new(http.Client),
		funcs:    make(template.FuncMap),
	}

	c.funcs["unix"] = unix
	c.funcs["nanoseconds"] = nanoseconds

	err = c.SetIndex()

	return
}

func (c InfluxdbCollector) SetIndex() (err error) {
	t := "CREATE DATABASE {{.Database}}"

	q, err := c.tmpl(t, c)
	if err != nil {
		return
	}

	return c.postQuery("query", q)
}

func (c InfluxdbCollector) Push(o loadtest.Output) (err error) {
	t := "request,url={{.URL}},method={{.Method}},status={{.Status}},error={{if .Error}}true{{else}}false{{end}} size={{.Size}},duration={{nanoseconds .Duration}} {{unix .Timestamp}}"
	q, err := c.tmpl(t, o)
	if err != nil {
		return
	}

	p := fmt.Sprintf("write?db=%s", c.Database)

	return c.post(p, q)
}

func (c InfluxdbCollector) tmpl(s string, i interface{}) (out string, err error) {
	tmpl, err := template.New("create").Funcs(c.funcs).Parse(s)
	if err != nil {
		return
	}

	outputBuffer := bytes.NewBuffer(make([]byte, 0))

	err = tmpl.Execute(outputBuffer, i)
	if err != nil {
		return
	}

	out = outputBuffer.String()

	return
}

func (c InfluxdbCollector) post(path, data string) (err error) {
	r := strings.NewReader(data)

	resp, err := c.client.Post(fmt.Sprintf("%s/%s", c.Host, path), "", r)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		if resp.Request == nil {
			err = fmt.Errorf("(INCOMPLETE LOG) %s - %s", resp.Status, string(body))
		} else {
			err = fmt.Errorf("%s on %s returned %s - %s",
				resp.Request.Method,
				resp.Request.URL.String(),
				resp.Status,
				string(body),
			)
		}
	}

	return
}

func (c InfluxdbCollector) postQuery(path, query string) (err error) {
	fmt.Println(query)

	resp, err := c.client.PostForm(fmt.Sprintf("%s/%s", c.Host, path), url.Values{"q": {query}})
	if err != nil {
		return
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

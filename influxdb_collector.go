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
)

type httpClient interface {
	Post(string, string, io.Reader) (*http.Response, error)
	PostForm(string, url.Values) (*http.Response, error)
}

// InfluxdbCollector is a collector which pushes into the popular timeseries
// database "Influx"
type InfluxdbCollector struct {
	Host     string
	Database string

	client  httpClient
	funcs   template.FuncMap
	indices map[string]byte
}

// NewInfluxdbCollector will bootstrap an InfluxdbCollector
// and configuree some `text/template' functions
func NewInfluxdbCollector(host, db string) (c InfluxdbCollector, err error) {
	c = InfluxdbCollector{
		Host:     host,
		Database: db,

		client:  new(http.Client),
		funcs:   make(template.FuncMap),
		indices: make(map[string]byte),
	}

	c.funcs["unix"] = unix
	c.funcs["nanoseconds"] = nanoseconds

	return
}

// CreateIndex takes a database name, templates it into an influx query,
// and posts this to influx. It then stores the index to ensure it's not
// created over and over. While this is idempotent in influx, it's still
// inefficient
func (c *InfluxdbCollector) CreateIndex(database string) (err error) {
	t := "CREATE DATABASE {{.}}"

	q, err := c.tmpl(t, database)
	if err != nil {
		return
	}

	err = c.postQuery("query", q)
	if err != nil {
		return
	}

	c.indices[database] = '1'

	return
}

// Push takes output and pushes it into an influx database. If this database
// is not known to exist (so: this instance of the collector doesn't know it)
// then it will create it
func (c InfluxdbCollector) Push(o OutputMapper) (err error) {
	if o.output.Timestamp.UnixNano() == -6795364578871345152 {
		// We've ingested some really invalid data that doesn't have even
		// a timestamp
		return fmt.Errorf("Missing timestamp")
	}

	t := "request,url={{.URL}},method={{.Method}},status={{.Status}},error={{if .Error}}true{{else}}false{{end}} size={{.Size}},duration={{nanoseconds .Duration}} {{unix .Timestamp}}"
	q, err := c.tmpl(t, o.output)
	if err != nil {
		return
	}

	if _, ok := c.indices[o.database]; !ok {
		err = c.CreateIndex(o.database)
		if err != nil {
			return
		}
	}

	p := fmt.Sprintf("write?db=%s", o.database)

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

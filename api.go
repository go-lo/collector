package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/jspc/loadtest"
)

type API struct {
	OutputChan chan OutputWriteWrapper
}

func (a API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/push/"):
		a.Push(w, r)

	default:
		http.Error(w, fmt.Sprintf("%s not found", r.URL.Path), http.StatusNotFound)
	}
}

func (a API) Push(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	index := strings.TrimPrefix(r.URL.Path, "/push/")

	o := new(loadtest.Output)

	err = json.Unmarshal(body, o)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	a.OutputChan <- OutputWriteWrapper{
		output:   *o,
		database: index,
	}
}

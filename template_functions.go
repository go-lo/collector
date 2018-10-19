package main

import (
	"net/url"
	"strconv"
	"time"
)

func unix(t time.Time) string {
	return strconv.Itoa(int(t.UnixNano()))
}

func nanoseconds(t time.Duration) int64 {
	return t.Nanoseconds()
}

func cleanURL(in string) (out string, err error) {
	u, err := url.Parse(in)
	if err != nil {
		return
	}

	u.RawQuery = ""
	u.User = nil

	out = u.String()

	return
}

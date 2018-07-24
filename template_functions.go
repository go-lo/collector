package main

import (
	"strconv"
	"time"
)

func unix(t time.Time) string {
	return strconv.Itoa(int(t.UnixNano()))
}

func nanoseconds(t time.Duration) int64 {
	return t.Nanoseconds()
}

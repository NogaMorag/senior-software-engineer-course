package platform

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Healthcheck is "shorty healthcheck" and "linkworker healthcheck". The
// distroless image has no shell and no curl, so a container healthcheck has to
// be a program that ships in the image: this one. Added in chapter 19 (in
// package main there) and moved here in chapter 24 so both commands share it.
//
// Healthcheck GETs url and returns the process exit code: 0 for a 200 answer,
// 1 for anything else. Errors go to w, so they show up in docker inspect.
func Healthcheck(url string, w io.Writer) int {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(w, "healthcheck:", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(w, "healthcheck:", url, "answered", resp.Status)
		return 1
	}
	return 0
}

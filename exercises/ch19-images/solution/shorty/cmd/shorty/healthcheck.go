package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// defaultHealthURL is where "shorty healthcheck" looks when given no URL. The
// distroless image has no shell and no curl, so a container healthcheck has to
// be a program that ships in the image: this one. Added for chapter 19.
const defaultHealthURL = "http://127.0.0.1:8080/healthz"

// healthcheck GETs url and returns the process exit code: 0 for a 200 answer,
// 1 for anything else. Errors go to w, so they show up in docker inspect.
func healthcheck(url string, w io.Writer) int {
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

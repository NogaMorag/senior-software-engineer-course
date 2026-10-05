// Command greeter greets visitors in their language and counts visits.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"example.com/ch27-review/internal/greet"
	"example.com/ch27-review/internal/visits"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	counter := visits.New()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /greet/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		counter.Add(name)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"greeting": greet.Greet(name, r.Header.Get("Accept-Language")),
			"visits":   counter.Report(name),
		})
	})

	addr := "127.0.0.1:8080"
	log.Info("listening", slog.String("addr", addr))
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("serve", slog.Any("err", err))
		os.Exit(1)
	}
}

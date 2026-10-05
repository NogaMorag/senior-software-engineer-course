// Command greeter greets visitors in their language and counts visits.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

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
		if err := json.NewEncoder(w).Encode(map[string]string{
			"greeting": greet.Greet(name, r.Header.Get("Accept-Language")),
			"visits":   counter.Report(name),
		}); err != nil {
			log.WarnContext(r.Context(), "write response", slog.Any("err", err))
		}
	})

	srv := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	log.Info("listening", slog.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil {
		log.Error("serve", slog.Any("err", err))
		os.Exit(1)
	}
}

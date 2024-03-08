package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	releaseCollector, err := NewReleaseCollector()
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(releaseCollector)

	go func() {
		for {
			releaseCollector.UpdateFromCatalog()
			time.Sleep(time.Minute * 30)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              ":8000",
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

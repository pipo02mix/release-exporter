package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var i = flag.String("update-cache-every", "0m", "The interval in minutes between updates (e.g., 10s, 2m, 1h)")

func main() {
	flag.Parse()

	interval, err := time.ParseDuration(*i)
	if err != nil {
		fmt.Println("Invalid 'update-cache-every' format. Please enter a duration string like 10s, 2m, or 1h.")
		log.Fatal(err)
	}

	releaseCollector, err := NewReleaseCollector()
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(releaseCollector)

	ticker := time.NewTicker(interval)

	go func() {
		for {
			fmt.Println("Updating entries from catalog.")
			releaseCollector.UpdateFromCatalog()
			<-ticker.C
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

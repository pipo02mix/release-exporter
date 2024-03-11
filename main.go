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

var interval = flag.Int("run-every", 30, "The interval in minutes between updates")

func main() {
	flag.Parse()

	releaseCollector, err := NewReleaseCollector()
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(releaseCollector)

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              ":8000",
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

	if *interval <= 0 {
		fmt.Println("No interval specified, running once and exiting.")
		releaseCollector.UpdateFromCatalog()
	} else {
		ticker := time.NewTicker(time.Duration(*interval) * time.Minute)

		go func() {
			for {
				releaseCollector.UpdateFromCatalog()
				<-ticker.C
			}
		}()
	}
}

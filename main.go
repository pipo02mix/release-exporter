package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var interval = flag.Int("run-every", 30, "The interval in minutes between updates")

func main() {
	flag.Parse()
	if *interval <= 0 {
		fmt.Println("The 'run-every' flag must be greater than 0.")
		os.Exit(0)
	}

	releaseCollector, err := NewReleaseCollector()
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(releaseCollector)

	ticker := time.NewTicker(time.Duration(*interval) * time.Minute)

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

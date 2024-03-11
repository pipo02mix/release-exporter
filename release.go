package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/Masterminds/semver/v3"
	"github.com/prometheus/client_golang/prometheus"
)

type Entry struct {
	Name        string            `yaml:"name"`
	Annotations map[string]string `yaml:"annotations"`
	Version     string            `yaml:"version"`
}

type Entries map[string][]Entry

type Catalog struct {
	APIVersion string  `yaml:"apiVersion"`
	Entries    Entries `yaml:"entries"`
}

type ComponentRelease struct {
	Name    string
	Release string
	Team    string
	State   float64
}

type releaseCollector struct {
	releaseDesc *prometheus.Desc

	componentReleases []ComponentRelease
}

func NewReleaseCollector() (*releaseCollector, error) {
	var rc releaseCollector

	rc.releaseDesc = prometheus.NewDesc(
		prometheus.BuildFQName("operations", "release", "latest_info"),
		"App release information from Giant Swarm Catalog index file.",
		[]string{"app", "team", "version"},
		nil,
	)

	rc.componentReleases = []ComponentRelease{}

	return &rc, nil
}

func (c *releaseCollector) UpdateFromCatalog() []ComponentRelease {
	resp, err := http.Get("https://raw.githubusercontent.com/giantswarm/giantswarm-catalog/master/index.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	catalog := &Catalog{}
	err = yaml.Unmarshal(data, &catalog)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var crs []ComponentRelease
	for _, entries := range catalog.Entries {
		sort.Slice(entries, func(i, j int) bool {
			versionI, _ := semver.NewVersion(entries[i].Version)
			versionJ, _ := semver.NewVersion(entries[j].Version)
			return versionI.GreaterThan(versionJ)
		})

		// Print the entry with the latest version
		if len(entries) > 0 {
			latest := entries[0]
			crs = append(crs, ComponentRelease{
				Name:    latest.Name,
				Release: latest.Version,
				Team:    latest.Annotations["application.giantswarm.io/team"],
				State:   1,
			})
			fmt.Printf("Latest release for %s: %s\n", latest.Name, latest.Version)
		}
	}

	return crs
}

func (c *releaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.releaseDesc
}

func (c *releaseCollector) Collect(ch chan<- prometheus.Metric) {
	for _, cr := range c.componentReleases {
		ch <- prometheus.MustNewConstMetric(
			c.releaseDesc,
			prometheus.GaugeValue,
			cr.State,
			cr.Name,
			cr.Team,
			cr.Release,
		)
	}
}

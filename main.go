package main

import (
	"github.com/neoll-labs/cosmos-block-fetcher/cmd"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	setupMetrics()
	cmd.Execute()
}

func setupMetrics() {
	// Add to your main HTTP server
	http.Handle("/metrics", promhttp.Handler())
}

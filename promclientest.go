package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var jobsInQueue = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "jobs_in_queue",
		Help: "Current number of jobs in the queue",
	},
)

func init() {
	prometheus.MustRegister(jobsInQueue)
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	panic(http.ListenAndServe(":8080", nil))

	//jobsInQueue.Dec()
	jobsInQueue.Inc()
	jobsInQueue.Inc()
	jobsInQueue.Inc()
	jobsInQueue.Inc()
}

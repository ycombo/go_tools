// Golang Prometheus monitor plugin for QPS and ResponseTime metrics
//
// For more information visit:
// https://freerecursion.wordpress.com/2017/12/17/%E4%BD%BF%E7%94%A8prometheus%E6%9D%A5%E7%9B%91%E6%8E%A7go-http-server/
//
// Usage:
/******************
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/ycombo/go_tools"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, there!")
}

func main() {
    handlerFunc := go_tools.CreateMonitorChain(map[string]bool{"all": true},
                                              helloHandler, true)
     http.Handle("/hello", handlerFunc)
     log.Fatal(http.ListenAndServe(":8010", nil))
}

*********************/

package go_tools

import (
    "net/http"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type monitorManager  struct {
    counterVec *prometheus.CounterVec
    histVec *prometheus.HistogramVec
    handlerFunc http.HandlerFunc
    metricsToMonitor map[string]bool
}

// Add the default '/metrics' url route path for Prometheus server to pull metrics data.
func init() {
    http.Handle("/metrics", promhttp.Handler())
}

// Create a HandlerFunc monitor chain wrapper to get QPS and repsonse time metrics
func CreateMonitorChain(mtm map[string]bool, hf http.HandlerFunc, useWrapper bool) http.HandlerFunc {
    if !useWrapper{
        return hf
    }

    mm := monitorManager{metricsToMonitor: mtm, handlerFunc: hf}
    if mm.metricsToMonitor["api_requests_total"] || mm.metricsToMonitor["all"] {
         mm.counterVec = prometheus.NewCounterVec(
             prometheus.CounterOpts{
                 Name: "api_requests_total",
                 Help: "A counter for requests to the wrapped handler.",
             },
             []string{"code", "method"},
         )

         prometheus.MustRegister(mm.counterVec)
    }

    if mm.metricsToMonitor["response_duration_seconds"] || mm.metricsToMonitor["all"] {
         mm.histVec = prometheus.NewHistogramVec(
             prometheus.HistogramOpts{
                 Name:        "response_duration_seconds",
                 Help:        "A histogram of request latencies.",
                 Buckets:     prometheus.DefBuckets,
                 ConstLabels: prometheus.Labels{"handler": "api"},
             },
             []string{"method"},

         )
         prometheus.MustRegister(mm.histVec)
    }

    return promhttp.InstrumentHandlerCounter(
               mm.counterVec, promhttp.InstrumentHandlerDuration(mm.histVec, mm.handlerFunc),)
}

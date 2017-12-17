package utils

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

func init() {
    http.Handle("/metrics", promhttp.Handler())
}

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

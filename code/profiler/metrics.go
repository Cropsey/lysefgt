package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	prometheusStats = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "profiler_aggregated",
		Help: "Counter for aggregated statistics of profiled application stack traces",
	}, []string{"pid", "symbol", "addr", "file", "line", "level", "self", "bin"})
)

func runPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	log.Fatalf("prometheus server failed: %v", err)
}

func addPrometheus(event bpfEvent, i int, ustack []stackPos) {
	pos := ustack[i]
	labels := []string{
		fmt.Sprintf("%d", event.Pid),
		pos.symbol,
		fmt.Sprintf("0x%x", pos.addr),
		pos.file,
		fmt.Sprintf("%d", pos.line),
		fmt.Sprintf("%d", i),
		fmt.Sprintf("%d", i),
		event.taskComm(),
	}
	prometheusStats.WithLabelValues(labels...).Inc()
}

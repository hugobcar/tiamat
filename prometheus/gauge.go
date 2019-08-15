package prometheus

import (
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	mu              *sync.Mutex
	RegistredGauges map[string]prometheus.Gauge
)

func init() {
	mu = new(sync.Mutex)
	RegistredGauges = make(map[string]prometheus.Gauge)
}

// CreateGauges - Create Gauges in Prometheus
func CreateGauges(queues []string) {
	for _, queue := range queues {
		queueSplit := strings.Split(queue, "/")
		queueRegionSplit := strings.Split(queueSplit[2], ".")
		queueRegion := queueRegionSplit[0]
		queueAccount := queueSplit[3]
		queueName := queueSplit[4]

		gID := fmt.Sprintf(queue)

		if g, found := RegistredGauges[gID]; found {
			fmt.Println(g)
		}

		g := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tiamat",
			Help: "Used to export SQS metrics",
			ConstLabels: prometheus.Labels{
				"namespace":     "namespace",
				"metric_type":   "SQS",
				"queue_region":  queueRegion,
				"queue_account": queueAccount,
				"queue_name":    queueName,
				"queue_url":     queue,
			},
		})
		prometheus.Register(g)
		RegistredGauges[gID] = g

	}
}

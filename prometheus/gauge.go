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
func CreateGauges(queues []string, formatGaugeName bool, metricType string) {
	for _, queue := range queues {
		queueSplit := strings.Split(queue, "/")
		queueRegionSplit := strings.Split(strings.ReplaceAll(queueSplit[2], "sqs.", ""), ".")
		queueRegion := queueRegionSplit[0]
		queueAccount := queueSplit[3]
		queueName := strings.ReplaceAll(queueSplit[4], "-", "_")

		gID := fmt.Sprintf(queue)

		if g, found := RegistredGauges[gID]; found {
			fmt.Println(g)
		}

		var gName = "tiamat"

		if formatGaugeName {
			gName = fmt.Sprintf("tiamat_%s_%s_%s", queueAccount, metricType, queueName)
		}

		g := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: gName,
			Help: "Used to export SQS metrics",
			ConstLabels: prometheus.Labels{
				"metric_type":   metricType,
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

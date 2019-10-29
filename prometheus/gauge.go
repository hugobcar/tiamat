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
		queueName := strings.ToLower(strings.ReplaceAll(queueSplit[4], "-", "_"))

		var gPrefix = "tiamat"
		if formatGaugeName {
			gPrefix = fmt.Sprintf("tiamat_%s_%s_%s", queueAccount, metricType, queueName)
		}

		gTotal := fmt.Sprintf("%s_total", gPrefix)
		gTotalName := fmt.Sprintf("%s_total", queue)

		gVisible := fmt.Sprintf("%s_visible", gPrefix)
		gVisibleName := fmt.Sprintf("%s_visible", queue)

		gInFlight := fmt.Sprintf("%s_in_flight", gPrefix)
		gInFlightName := fmt.Sprintf("%s_in_flight", queue)

		defaultLabels := prometheus.Labels{
			"metric_type":   metricType,
			"queue_region":  queueRegion,
			"queue_account": queueAccount,
			"queue_name":    queueName,
			"queue_url":     queue,
		}

		registerGuage(gTotalName,gTotal, defaultLabels, "SQS Total Messages metrics")
		registerGuage(gVisibleName,gVisible, defaultLabels, "SQS Visible Messages metrics")
		registerGuage(gInFlightName,gInFlight, defaultLabels, "SQS In Fight Messages metrics")
	}
}

func registerGuage(name , metric string, labels prometheus.Labels, help string)  {
	gID := fmt.Sprintf(name)

	if g, found := RegistredGauges[gID]; found {
		fmt.Println(g)
	}

	g := prometheus.NewGauge(prometheus.GaugeOpts{Name: metric, Help: help, ConstLabels: labels})
	prometheus.Register(g)
	RegistredGauges[gID] = g
}

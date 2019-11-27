package prometheus

import (
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type QueueMetrics struct {
	Name     string
	Total    prometheus.Gauge
	Visible  prometheus.Gauge
	InFlight prometheus.Gauge
	Legacy   prometheus.Gauge
}

const (
	Total    = "total"
	Visible  = "visible"
	InFlight = "in_flight"
)

var (
	mu              *sync.Mutex
	RegistredGauges map[string]QueueMetrics
)

func init() {
	mu = new(sync.Mutex)
	RegistredGauges = make(map[string]QueueMetrics)
}

// CreateGauges - Create Gauges in Prometheus
func CreateGauges(queues []string, formatGaugeName bool, metricType string) {
	for _, queue := range queues {
		queueSplit := strings.Split(queue, "/")
		queueRegionSplit := strings.Split(strings.ReplaceAll(queueSplit[2], "sqs.", ""), ".")
		queueRegion := queueRegionSplit[0]
		queueAccount := queueSplit[3]
		queueName := strings.ToLower(strings.ReplaceAll(queueSplit[4], "-", "_"))

		var gName = "tiamat"
		if formatGaugeName {
			gName = fmt.Sprintf("tiamat_%s_%s_%s", queueAccount, metricType, queueName)
		}

		defaultLabels := prometheus.Labels{
			"metric_type":   metricType,
			"queue_region":  queueRegion,
			"queue_account": queueAccount,
			"queue_name":    queueName,
			"queue_url":     queue,
		}

		if g, found := RegistredGauges[queue]; found {
			fmt.Println(g)
		}

		RegistredGauges[queue] = QueueMetrics{
			Name:     gName,
			Total:    RegisterGauge(gName, fmt.Sprintf("%s_%s", gName, Total), defaultLabels, "SQS Total Messages metrics"),
			Visible:  RegisterGauge(gName, fmt.Sprintf("%s_%s", gName, Visible), defaultLabels, "SQS Visible Messages metrics"),
			InFlight: RegisterGauge(gName, fmt.Sprintf("%s_%s", gName, InFlight), defaultLabels, "SQS In Fight Messages metrics"),
			Legacy:   RegisterGauge(gName, gName, defaultLabels, "Legacy metric, use total metrics instead"),
		}
	}
}

// RegisterGauge -- register new prometheus guage metrics
func RegisterGauge(name, metric string, labels prometheus.Labels, help string) prometheus.Gauge {
	if g, found := RegistredGauges[name]; found {
		fmt.Println(g)
	}

	g := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        metric,
			Help:        help,
			ConstLabels: labels,
		})

	if err := prometheus.Register(g); err != nil {
		fmt.Println(err)
	}
	return g
}

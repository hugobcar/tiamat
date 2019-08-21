package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hugobcar/tiamat/aws"
	"github.com/hugobcar/tiamat/prometheus"
)

var s *aws.SQSColector
var wg sync.WaitGroup

type result struct {
	queueName string
	totalMsg  float64
}

type config struct {
	Region          string   `json:"region"`
	Queues          []string `json:"queue_urls"`
	Interval        int      `json:"interval"`
	FormatGaugeName bool     `json:"format_gauge_name"`
	MetricType      string   `json:"metric_type"`
}

func checkEmptyVariable(name, variable string) {
	if variable == "0" || len(strings.TrimSpace(variable)) == 0 {
		fmt.Printf("Please set %s", name)

		os.Exit(2)
	}
}

func validMetricType(variable string) {
	if variable != "sqs" {
		fmt.Printf("Metric type (%s) is invalid!", variable)

		os.Exit(2)
	}
}

func main() {
	var configStruct config

	file, _ := ioutil.ReadFile("config.json")
	_ = json.Unmarshal([]byte(file), &configStruct)

	awsKey := os.Getenv("AWSKEY")
	awsSecret := os.Getenv("AWSSECRET")
	awsRegion := configStruct.Region
	queues := configStruct.Queues
	interval := configStruct.Interval
	formatGaugeName := configStruct.FormatGaugeName
	metricType := configStruct.MetricType

	// Test empty confs variables
	checkEmptyVariable("secret AWSKEY", awsKey)
	checkEmptyVariable("secret AWSSECRET", awsSecret)
	checkEmptyVariable("configMap value: region", awsRegion)
	checkEmptyVariable("configMap value: queue_urls", strconv.Itoa(len(queues)))
	checkEmptyVariable("configMap value: interval", strconv.Itoa(interval))
	checkEmptyVariable("configMap value: metric_type", metricType)

	// Check Metric type
	validMetricType(metricType)

	go prometheus.Run()
	go prometheus.CreateGauges(queues, formatGaugeName, metricType)

	r := make(chan result)

	for {
		wg.Add(len(queues))

		var ini time.Time
		ini = time.Now()

		for _, url := range queues {
			go run(awsKey, awsSecret, awsRegion, url, r)
		}

		wg.Wait()

		fmt.Println("(Duration time to get total messages in SQS: ", time.Since(ini).Seconds(), "seconds)")

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func run(awsKey, awsSecret, awsRegion, url string, rchan chan result) {
	defer close(rchan)

	var r result

	t, err := s.GetMetrics(awsKey, awsSecret, awsRegion, url)
	if err != nil {
		t = -1
	}

	r.totalMsg = t
	r.queueName = url

	prometheus.RegistredGauges[url].Set(t)

	wg.Done()
	rchan <- r
}

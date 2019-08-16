package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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
	Region   string   `json:"region"`
	Queues   []string `json:"queue_urls"`
	Interval string   `json:"interval"`
}

func main() {
	var configStruct config

	file, _ := ioutil.ReadFile("config.json")
	_ = json.Unmarshal([]byte(file), &configStruct)

	awsKey := os.Getenv("AWSKEY")
	awsSecret := os.Getenv("AWSSECRET")
	awsRegion := configStruct.Region
	queues := configStruct.Queues

	interval, err := strconv.Atoi(configStruct.Interval)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	go prometheus.Run()
	go prometheus.CreateGauges(queues)

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

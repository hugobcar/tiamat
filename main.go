package main

import (
	"os"
	"time"

	"github.com/hugobcar/tiamat/aws"
	"github.com/hugobcar/tiamat/prometheus"
)

var s *aws.SQSColector

type result struct {
	queueName string
	totalMsg  float64
}

func main() {
	queues := []string{
		"https://sqs.sa-east-1.amazonaws.com/739171219021/teste_ana",
		"https://sqs.sa-east-1.amazonaws.com/739171219021/testehugo",
		// "https://queue.amazonaws.com/DEV_IFOOD_PAYMENT_RECONCILIATION_MANAGE_OPS",
	}
	interval := 5

	awsKey := os.Getenv("AWSKEY")
	awsSecret := os.Getenv("AWSSECRET")
	awsRegion := os.Getenv("AWSREGION")

	go prometheus.Run()
	go prometheus.CreateGauges(queues)

	r := make(chan result)

	for {
		for _, url := range queues {
			go run(awsKey, awsSecret, awsRegion, url, r)
		}

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

	rchan <- r
}

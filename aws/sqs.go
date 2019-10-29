package aws

import (
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	numberOfMessagesInQueueAttrName       = "ApproximateNumberOfMessages"
	numberOfMessagesInFlightQueueAttrName = "ApproximateNumberOfMessagesNotVisible"
	msgsInQueueMetricName                 = "msgsInQueue"
)

// SQSColector - SQSClient struct using pointer
type SQSColector struct {
	cli *SQSClient
}

// SQSClient - AWS struct
type SQSClient struct {
	*AWS
}

type SQSMetrics struct {
	Visible Metric
	InFlight Metric
}

type Metric struct {
	Name  string
	Value int
}

func (m SQSMetrics) TotalMessages() int {
	return m.InFlight.Value + m.Visible.Value
}

// GetMetrics - Used to get metrics (number msgs in queue) in SQS
func (s *SQSColector) GetMetrics(awsKey, awsSecret, awsRegion, queue string, logs bool) (SQSMetrics, error) {
	metrics, err := s.getNumberOfMsgsInQueue(awsKey, awsSecret, awsRegion, queue)
	if err != nil {
		return metrics, err
	}

	if logs {
		log.Printf("Messages in Queue (%s): Visible: %d In Flight: %d Total: %d\n",
			queue, metrics.Visible.Value, metrics.InFlight.Value, metrics.TotalMessages())
	}

	return metrics, nil
}

func (s *SQSColector) getNumberOfMsgsInQueue(awsKey, awsSecret, awsRegion, queueURL string) (SQSMetrics, error) {
	c := newSQSClient(awsKey, awsSecret, awsRegion)
	metrics := SQSMetrics{}

	attrs, err := c.getQueueAttributes(
		queueURL,
		numberOfMessagesInQueueAttrName,
		numberOfMessagesInFlightQueueAttrName,
	)
	if err != nil {
		return metrics, err
	}

	metrics.Visible.Value, err = readIntAttribute(attrs, numberOfMessagesInQueueAttrName)
	if err != nil {
		return metrics, err
	}

	metrics.InFlight.Value, err = readIntAttribute(attrs, numberOfMessagesInFlightQueueAttrName)
	if err != nil {
		return metrics, err
	}
	return metrics, nil
}

func readIntAttribute(attributes map[string]string, key string) (int, error) {
	m, err := strconv.Atoi(attributes[key])
	if err != nil {
		return m, err
	}
	return m, nil
}

func (s *SQSClient) getQueueAttributes(queueURL string, attributes ...string) (map[string]string, error) {
	cli := sqs.New(session.New(), s.newConfig())

	var attrList []*string
	for _, attr := range attributes {
		a := attr
		attrList = append(attrList, &a)
	}

	out, err := cli.GetQueueAttributes(
		&sqs.GetQueueAttributesInput{QueueUrl: &queueURL, AttributeNames: attrList},
	)

	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range out.Attributes {
		result[k] = *v
	}
	return result, nil
}

func newSQSClient(key, secret, region string) *SQSClient {
	return &SQSClient{
		AWS: &AWS{key: key, secret: secret, region: region},
	}
}

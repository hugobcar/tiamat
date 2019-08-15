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

// GetMetrics - Used to get metrics (number msgs in queue) in SQS
func (s *SQSColector) GetMetrics(awsKey, awsSecret, awsRegion, queue string) (float64, error) {
	n, err := s.getNumberOfMsgsInQueue(awsKey, awsSecret, awsRegion, queue)
	if err != nil {
		return 0, err
	}

	log.Printf("Messages in Queue (%s): %d\n", queue, n)

	return float64(n), nil
}

func (s *SQSColector) getNumberOfMsgsInQueue(awsKey, awsSecret, awsRegion, queueURL string) (int, error) {
	c := newSQSClient(awsKey, awsSecret, awsRegion)

	attrs, err := c.getQueueAttributes(
		queueURL,
		numberOfMessagesInQueueAttrName,
		numberOfMessagesInFlightQueueAttrName,
	)
	if err != nil {
		return -1, err
	}
	visible, err := strconv.Atoi(attrs[numberOfMessagesInQueueAttrName])
	if err != nil {
		return -1, err
	}
	inFlight, err := strconv.Atoi(attrs[numberOfMessagesInFlightQueueAttrName])
	if err != nil {
		return -1, err
	}
	return visible + inFlight, nil
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

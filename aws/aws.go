package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type AWS struct {
	key    string
	secret string
	region string
}

func (a *AWS) newConfig() *aws.Config {
	return aws.NewConfig().WithCredentials(
		credentials.NewStaticCredentials(a.key, a.secret, ""),
	).WithRegion(a.region)
}

package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// AWS - AWS parameters
type AWS struct {
	key    string
	secret string
	region string
}

func (a *AWS) newConfig() *aws.Config {
	var config *aws.Config

	if a.key == "" {
		config = aws.NewConfig()
	} else {
		config = aws.NewConfig().WithCredentials(
			credentials.NewStaticCredentials(a.key, a.secret, ""),
		)
	}

	return config.WithRegion(a.region)
}

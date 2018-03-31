package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/marksauter/markus-ninja-api/pkg/utils"
)

var AwsRegion = utils.GetOptionalEnv("AWS_REGION", "us-east-1")

func NewSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AwsRegion),
	})

	if err != nil {
		panic(err)
	}

	return sess
}

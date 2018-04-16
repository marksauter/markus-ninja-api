package myaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

var AwsRegion = util.GetOptionalEnv("AWS_REGION", "us-east-1")
var AwsSession = NewSession()

func NewSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AwsRegion),
	})

	if err != nil {
		panic(err)
	}

	return sess
}

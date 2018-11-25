package myaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

var AWSRegion = util.GetOptionalEnv("AWS_REGION", "us-east-1")
var AWSSession = NewSession()

func NewSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AWSRegion),
	})

	if err != nil {
		panic(err)
	}

	return sess
}

func GetSession() *session.Session {
	return AWSSession
}

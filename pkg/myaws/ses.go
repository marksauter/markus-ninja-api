package myaws

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
)

func NewSES() *ses.SES {
	return ses.New(AWSSession)
}

type MockSES struct {
	sesiface.SESAPI
}

func NewMockSES() *MockSES {
	return new(MockSES)
}

var MockSESServiceError = false

func (m *MockSES) SendEmail(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	output := new(ses.SendEmailOutput)
	if MockSESServiceError {
		return output, errors.New("AwsError")
	}
	return output, nil
}

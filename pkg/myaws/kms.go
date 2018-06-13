package myaws

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

func NewKMS() *kms.KMS {
	return kms.New(AWSSession)
}

type MockKMS struct {
	kmsiface.KMSAPI
}

func NewMockKMS() *MockKMS {
	return new(MockKMS)
}

var MockKMSServiceError = false

func (m *MockKMS) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	output := new(kms.EncryptOutput)
	if MockKMSServiceError {
		return output, errors.New("AwsError")
	}
	return output.SetCiphertextBlob(input.Plaintext), nil
}

func (m *MockKMS) Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error) {
	output := new(kms.DecryptOutput)
	if MockKMSServiceError {
		return output, errors.New("AwsError")
	}
	return output.SetPlaintext(input.CiphertextBlob), nil
}

package myaws

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

func NewKms() *kms.KMS {
	return kms.New(AwsSession)
}

type MockKms struct {
	kmsiface.KMSAPI
}

func NewMockKms() *MockKms {
	return new(MockKms)
}

var MockKmsServiceError = false

func (m *MockKms) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	output := new(kms.EncryptOutput)
	if MockKmsServiceError {
		return output, errors.New("AwsError")
	}
	return output.SetCiphertextBlob(input.Plaintext), nil
}

func (m *MockKms) Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error) {
	output := new(kms.DecryptOutput)
	if MockKmsServiceError {
		return output, errors.New("AwsError")
	}
	return output.SetPlaintext(input.CiphertextBlob), nil
}

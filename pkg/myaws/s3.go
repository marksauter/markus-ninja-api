package myaws

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

func NewS3() *s3.S3 {
	return s3.New(AWSSession)
}

type MockS3 struct {
	s3iface.S3API
}

func NewMockS3() *MockS3 {
	return new(MockS3)
}

var MockS3ServiceError = false

func (m *MockS3) AbortMultipartUpload(
	input *s3.AbortMultipartUploadInput,
) (*s3.AbortMultipartUploadOutput, error) {
	output := new(s3.AbortMultipartUploadOutput)
	if MockS3ServiceError {
		return output, errors.New("AwsError")
	}
	return output, nil
}

func (m *MockS3) CompleteMultipartUpload(
	input *s3.CompleteMultipartUploadInput,
) (*s3.CompleteMultipartUploadOutput, error) {
	output := new(s3.CompleteMultipartUploadOutput)
	if MockS3ServiceError {
		return output, errors.New("AwsError")
	}
	return output, nil
}

func (m *MockS3) CreateMultipartUpload(
	input *s3.CreateMultipartUploadInput,
) (*s3.CreateMultipartUploadOutput, error) {
	output := new(s3.CreateMultipartUploadOutput)
	if MockS3ServiceError {
		return output, errors.New("AwsError")
	}
	return output, nil
}

func (m *MockS3) UploadPart(
	input *s3.UploadPartInput,
) (*s3.UploadPartOutput, error) {
	output := new(s3.UploadPartOutput)
	if MockS3ServiceError {
		return output, errors.New("AwsError")
	}
	return output, nil
}

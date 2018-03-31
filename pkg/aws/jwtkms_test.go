package aws_test

import (
	"testing"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/aws"
	"github.com/marksauter/markus-ninja-api/pkg/jwt"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

type mockKMSClient struct {
	kmsiface.KMSAPI
}

var testPayload = jwt.NewPayload(&jwt.NewPayloadInput{
	Iat: time.Now(),
	Exp: time.Now().Add(time.Minute * time.Duration(10)),
	Id:  "asdf",
})
var testToken = jwt.Token{Payload: testPayload, Signature: "signature"}

var signature = "ciphertext"

func (m *mockKMSClient) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	var output kms.EncryptOutput
	ciphertextBlob := []byte(signature)
	return output.SetCiphertextBlob(ciphertextBlob), nil
}

var plaintext = "plaintext"

func (m *mockKMSClient) Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error) {
	var output kms.DecryptOutput
	return output.SetPlaintext([]byte(plaintext)), nil
}

var mockJwtKms = aws.JwtKms{&mockKMSClient{}}

func TestEncrypt(t *testing.T) {
	expected := jwt.Token{Payload: testPayload, Signature: signature}
	actual := mockJwtKms.Encrypt(testPayload)
	if actual != expected {
		t.Errorf(
			"TestEncrypt(%+v): expected %s, actual %s",
			testPayload,
			expected,
			actual,
		)
	}
}

func TestDecrypt(t *testing.T) {
	expected := testPayload
	actual, _ := mockJwtKms.Decrypt(testToken.String())
	if actual != expected {
		t.Errorf(
			"TestDecrypt(%s): expected %#v, actual %#v",
			testToken.String(),
			expected,
			actual,
		)
	}
}

package aws

import (
	"errors"
	"fmt"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/jwt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

type JwtKms struct {
	Svc kmsiface.KMSAPI
}

// sess := NewSession()
// kms.New(sess)

var keyId = "alias/markus-ninja-api-key-alias"

func (jk *JwtKms) Encrypt(p jwt.Payload) jwt.Token {
	token := jwt.Token{Payload: p}

	params := &kms.EncryptInput{
		KeyId:     aws.String(keyId),
		Plaintext: []byte(token.GetPlainText()),
	}

	result, err := jk.Svc.Encrypt(params)
	if err != nil {
		panic(err)
	}

	token.Signature = string(result.CiphertextBlob)
	return token
}

func (jk *JwtKms) Decrypt(token string) (jwt.Payload, error) {
	t := jwt.ParseTokenString(token)

	now := time.Now()

	// Allow for server times that are 10 mins ahead of the local time
	issuedAt := t.Payload.IssuedAt().Add(-time.Minute * time.Duration(10))

	if issuedAt.After(now) {
		return jwt.Payload{}, errors.New("Token was issued after the current time")
	}

	if t.Payload.ExpiresAt().Before(now) {
		return jwt.Payload{}, errors.New("Token is expired")
	}

	params := &kms.DecryptInput{CiphertextBlob: []byte(t.Signature)}

	result, err := jk.Svc.Decrypt(params)
	if err != nil {
		return jwt.Payload{}, errors.New(fmt.Sprintf("AWS Error: %s", err))
	}

	if t.GetPlainText() != string(result.Plaintext) {
		return jwt.Payload{}, errors.New("Invalid signature")
	}

	return t.Payload, nil
}

package myaws

import (
	"errors"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/jwt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

var ErrTokenExpired = errors.New("Token is expired")
var ErrTokenIssuedAfter = errors.New("Token was issued after the current time")
var ErrInvalidSignature = errors.New("Invalid signature")

type JwtKms struct {
	keyId string
	Svc   kmsiface.KMSAPI
}

func NewJwtKms() *JwtKms {
	return &JwtKms{
		keyId: "alias/markus-ninja-api-key-alias",
		Svc:   NewKms(),
	}
}

func NewMockJwtKms() *JwtKms {
	return &JwtKms{
		keyId: "alias/markus-ninja-api-key-alias",
		Svc:   NewMockKms(),
	}
}

func (jk *JwtKms) Encrypt(p *jwt.Payload) *jwt.Token {
	token := jwt.Token{Payload: *p}

	params := &kms.EncryptInput{
		KeyId:     aws.String(jk.keyId),
		Plaintext: []byte(token.GetPlainText()),
	}

	result, err := jk.Svc.Encrypt(params)
	if err != nil {
		panic(err)
	}

	token.Signature = string(result.CiphertextBlob)
	return &token
}

func (jk *JwtKms) Decrypt(t *jwt.Token) (*jwt.Payload, error) {
	now := time.Now()

	// Allow for server times that are 10 mins ahead of the local time
	issuedAt := t.Payload.IssuedAt().Add(-time.Minute * time.Duration(10))

	if issuedAt.After(now) {
		return new(jwt.Payload), ErrTokenIssuedAfter
	}

	expiresAt := t.Payload.ExpiresAt()
	if !expiresAt.IsZero() && expiresAt.Before(now) {
		return new(jwt.Payload), ErrTokenExpired
	}

	params := &kms.DecryptInput{CiphertextBlob: []byte(t.Signature)}

	result, err := jk.Svc.Decrypt(params)
	if err != nil {
		panic(err)
	}

	if t.GetPlainText() != string(result.Plaintext) {
		return new(jwt.Payload), ErrInvalidSignature
	}

	payload := t.Payload

	return &payload, nil
}

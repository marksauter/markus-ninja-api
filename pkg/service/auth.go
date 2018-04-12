package service

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

type AuthService struct {
	keyId  string
	logger *mylog.Logger
	svc    kmsiface.KMSAPI
}

func NewAuthService(logger *mylog.Logger) *AuthService {
	return &AuthService{
		keyId:  "alias/markus-ninja-api-key-alias",
		logger: logger,
		svc:    myaws.NewKms(),
	}
}

func NewMockAuthService() *AuthService {
	return &AuthService{
		keyId: "alias/markus-ninja-api-key-alias",
		svc:   myaws.NewMockKms(),
	}
}

func (s *AuthService) SignJWT(p *myjwt.Payload) (*myjwt.JWT, error) {
	jwt := myjwt.JWT{Payload: *p}

	params := &kms.EncryptInput{
		KeyId:     aws.String(s.keyId),
		Plaintext: []byte(jwt.GetPlainText()),
	}

	result, err := s.svc.Encrypt(params)
	if err != nil {
		return nil, err
	}

	jwt.Signature = string(result.CiphertextBlob)
	return &jwt, nil
}

func (s *AuthService) ValidateJWT(t *myjwt.JWT) (*myjwt.Payload, error) {
	now := time.Now()

	// Allow for server times that are 10 mins ahead of the local time
	issuedAt := time.Unix(t.Payload.Iat, 0).Add(-time.Minute * time.Duration(10))

	if issuedAt.After(now) {
		return new(myjwt.Payload), myjwt.ErrTokenIssuedAfter
	}

	expiresAt := time.Unix(t.Payload.Exp, 0)
	if !expiresAt.IsZero() && expiresAt.Before(now) {
		return new(myjwt.Payload), myjwt.ErrTokenExpired
	}

	params := &kms.DecryptInput{CiphertextBlob: []byte(t.Signature)}

	result, err := s.svc.Decrypt(params)
	if err != nil {
		panic(err)
	}

	if t.GetPlainText() != string(result.Plaintext) {
		return new(myjwt.Payload), myjwt.ErrInvalidSignature
	}

	payload := t.Payload

	return &payload, nil
}

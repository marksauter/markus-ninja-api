package myjwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type UnixTimestamp = int64
type JWTErrorCode int

const (
	UnknownJWTError JWTErrorCode = iota
	InvalidToken
	InvalidSignature
	TokenExpired
	InvalidScope
	UnauthorizedClient
	UnsupportedGrantType
)

var ErrTokenExpired = errors.New("The access token provided has expired")
var ErrTokenIssuedAfter = errors.New("Token was issued after the current time")
var ErrInvalidSignature = errors.New("Invalid signature")

type Payload struct {
	// The id of the user for which the token was released (Subject)
	Sub string
	// UNIX timestamp when the token expires (Expiration)
	Exp UnixTimestamp
	// UNIX timestamp when the token was created (Issued At)
	Iat UnixTimestamp
	// Space-separated list of scopes for which the token is issued
	Scope string
}

func (p Payload) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(data)
}

// Modified version of a json web token without a header
type JWT struct {
	Payload   Payload
	Signature string
}

func (t JWT) String() string {
	return fmt.Sprintf("%v.%v", t.Payload, url.QueryEscape(t.Signature))
}

func (t JWT) GetPlainText() string {
	return t.Payload.String()
}

var ErrInvalidJWTFormat = errors.New(`Invalid JWT: expected format "payload.signature"`)
var ErrInvalidJWTEncoding = errors.New("Invalid JWT: expected base64 encoded")
var ErrInvalidJWTPayload = errors.New("Invalid JWT Payload")
var ErrInvalidJWTSignature = errors.New("Invalid JWT Signature")

func ParseToken(token string) (*JWT, error) {
	components := strings.SplitN(token, ".", 2)
	if len(components) != 2 {
		return new(JWT), ErrInvalidJWTFormat
	}
	decodedPayload, err := base64.URLEncoding.DecodeString(components[0])
	if err != nil {
		return new(JWT), ErrInvalidJWTEncoding
	}

	payload := new(Payload)

	err = json.Unmarshal(decodedPayload, payload)
	if err != nil {
		return new(JWT), ErrInvalidJWTPayload
	}

	signature, err := url.QueryUnescape(components[1])
	if err != nil {
		return new(JWT), ErrInvalidSignature
	}
	return &JWT{Payload: *payload, Signature: signature}, nil
}

package jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Payload struct {
	// RFC3339 formatted time string, identifies when token expires.
	Exp string
	// RFC3339 formatted time string, identifies when token was issued.
	Iat string
	// User ID
	Id string
}

func (p Payload) ExpiresAt() time.Time {
	var t time.Time
	var err error
	if p.Exp != "" {
		t, err = time.Parse(time.RFC3339, p.Exp)
	}
	if err != nil {
		panic(err)
	}
	return t
}

func (p Payload) IssuedAt() time.Time {
	var t time.Time
	var err error
	if p.Iat != "" {
		t, err = time.Parse(time.RFC3339, p.Iat)
	}
	if err != nil {
		panic(err)
	}
	return t
}

func (p Payload) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(data)
}

type NewPayloadInput struct {
	Exp time.Time
	Iat time.Time
	Id  string
}

func NewPayload(input *NewPayloadInput) Payload {
	exp := input.Exp.Format(time.RFC3339)
	iat := input.Iat.Format(time.RFC3339)
	return Payload{Id: input.Id, Exp: exp, Iat: iat}
}

// Modified version of a json web token without a header
type Token struct {
	Payload   Payload
	Signature string
}

func (t Token) String() string {
	return fmt.Sprintf("%v.%v", t.Payload, t.Signature)
}

func (t Token) GetPlainText() string {
	return t.Payload.String()
}

var ErrInvalidTokenFormat = errors.New(`Invalid token: expected format "payload.signature"`)
var ErrInvalidTokenEncoding = errors.New("Invalid token: expected base64 encoded")
var ErrInvalidTokenPayload = errors.New("Invalid Token Payload")

func ParseTokenString(token string) (*Token, error) {
	components := strings.Split(token, ".")
	if len(components) != 2 {
		return new(Token), ErrInvalidTokenFormat
	}
	decodedPayload, err := base64.URLEncoding.DecodeString(components[0])
	if err != nil {
		return new(Token), ErrInvalidTokenEncoding
	}

	p := new(Payload)

	err = json.Unmarshal(decodedPayload, p)
	if err != nil {
		return new(Token), ErrInvalidTokenPayload
	}

	return &Token{Payload: *p, Signature: components[1]}, nil
}

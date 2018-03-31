package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type Payload struct {
	// RFC3339 formatted time string, identifies when token expires.
	Exp string
	// RFC3339 formatted time string, identifies when token was issued.
	Iat string
	Id  string
}

func (p Payload) ExpiresAt() time.Time {
	t, err := time.Parse(time.RFC3339, p.Exp)
	if err != nil {
		panic(err)
	}
	return t
}

func (p Payload) IssuedAt() time.Time {
	t, err := time.Parse(time.RFC3339, p.Iat)
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

func ParseTokenString(token string) Token {
	components := strings.Split(token, ".")
	decodedPayload, err := base64.URLEncoding.DecodeString(components[0])
	if err != nil {
		panic(err)
	}

	p := Payload{}

	err = json.Unmarshal(decodedPayload, &p)
	if err != nil {
		log.Fatal("Invalid Token Payload")
		panic(err)
	}

	s := components[1]

	return Token{Payload: p, Signature: s}
}

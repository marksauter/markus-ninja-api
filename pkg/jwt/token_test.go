package jwt_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/jwt"
)

var testPayload = jwt.NewPayload(&jwt.NewPayloadInput{
	Iat: time.Now(),
	Exp: time.Now().Add(time.Minute * time.Duration(10)),
	Id:  "asdf",
})
var testToken = jwt.Token{Payload: testPayload, Signature: "signature"}

func TestPayloadString(t *testing.T) {
	ps, _ := json.Marshal(testPayload)

	expected := base64.URLEncoding.EncodeToString([]byte(ps))
	actual := testPayload.String()
	if actual != expected {
		t.Errorf(
			"TestPayloadString(): expected %s, actual %s",
			expected,
			actual,
		)
	}
}

func TestTokenString(t *testing.T) {
	expected := fmt.Sprintf("%v.%v", testToken.Payload, testToken.Signature)
	actual := testToken.String()
	if actual != expected {
		t.Errorf(
			"TestTokenString(): expected %s, actual %s",
			expected,
			actual,
		)
	}
}

func TestGetPlainText(t *testing.T) {
	expected := testToken.Payload.String()
	actual := testToken.GetPlainText()
	if actual != expected {
		t.Errorf(
			"TestGetPlainText(): expected %s, actual %s",
			expected,
			actual,
		)
	}
}

func TestParseTokenString(t *testing.T) {
	expected := testToken
	actual := jwt.ParseTokenString(testToken.String())
	if actual != expected {
		t.Errorf(
			"TestParseTokenString(): expected %#v, actual %#v",
			expected,
			actual,
		)
	}
}

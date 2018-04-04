package jwt_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/jwt"
)

var mockPayload = jwt.NewPayload(&jwt.NewPayloadInput{
	Iat: time.Now(),
	Exp: time.Now().Add(time.Minute * time.Duration(10)),
	Id:  "asdf",
})
var mockToken = jwt.Token{Payload: mockPayload, Signature: mockPayload.String()}
var mockTokenString = base64.URLEncoding.EncodeToString([]byte("foo")) +
	"." +
	base64.URLEncoding.EncodeToString([]byte("bar"))

func TestPayloadString(t *testing.T) {
	ps, _ := json.Marshal(mockPayload)

	expected := base64.URLEncoding.EncodeToString([]byte(ps))
	actual := mockPayload.String()
	if actual != expected {
		t.Errorf(
			"TestPayloadString(): expected %s, actual %s",
			expected,
			actual,
		)
	}
}

func TestTokenString(t *testing.T) {
	expected := fmt.Sprintf("%v.%v", mockToken.Payload, mockToken.Signature)
	actual := mockToken.String()
	if actual != expected {
		t.Errorf(
			"TestTokenString(): expected %s, actual %s",
			expected,
			actual,
		)
	}
}

func TestGetPlainText(t *testing.T) {
	expected := mockToken.Payload.String()
	actual := mockToken.GetPlainText()
	if actual != expected {
		t.Errorf(
			"TestGetPlainText(): expected %s, actual %s",
			expected,
			actual,
		)
	}
}

func TestParseTokenStringSuccess(t *testing.T) {
	expected := mockToken
	actual, _ := jwt.ParseTokenString(mockToken.String())
	if *actual != expected {
		t.Errorf(
			"TestParseTokenString(): expected %#v, actual %#v",
			expected,
			actual,
		)
	}
}

var parseTokenStringFailureTests = []struct {
	t        string
	expected error
}{
	// Token expected to be headless with format "payload.signature"
	{"", jwt.ErrInvalidTokenFormat},
	{"foo.bar.baz", jwt.ErrInvalidTokenFormat},
	{"foo.bar", jwt.ErrInvalidTokenEncoding},
	{mockTokenString, jwt.ErrInvalidTokenPayload},
}

func TestParseTokenStringFailure(t *testing.T) {
	for _, tt := range parseTokenStringFailureTests {
		_, actual := jwt.ParseTokenString(tt.t)
		if actual != tt.expected {
			t.Errorf(
				"TestParseTokenString(%s): expected %#v, actual %#v",
				tt.t,
				tt.expected,
				actual,
			)
		}
	}
}

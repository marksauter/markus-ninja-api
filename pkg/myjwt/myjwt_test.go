package myjwt_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
)

var mockPayload = myjwt.Payload{
	Iat: time.Now().Unix(),
	Exp: time.Now().Add(time.Minute * time.Duration(10)).Unix(),
	Sub: "asdf",
}
var mockJWT = myjwt.JWT{Payload: mockPayload, Signature: mockPayload.String()}
var mockJWTString = base64.URLEncoding.EncodeToString([]byte("foo")) +
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

func TestJWTString(t *testing.T) {
	expected := fmt.Sprintf("%v.%v", mockJWT.Payload, mockJWT.Signature)
	actual := mockJWT.String()
	if actual != expected {
		t.Errorf(
			"TestJWTString(): expected %s, actual %s",
			expected,
			actual,
		)
	}
}

func TestGetPlainText(t *testing.T) {
	expected := mockJWT.Payload.String()
	actual := mockJWT.GetPlainText()
	if actual != expected {
		t.Errorf(
			"TestGetPlainText(): expected %s, actual %s",
			expected,
			actual,
		)
	}
}

func TestParseTokenSuccess(t *testing.T) {
	expected := mockJWT
	actual, _ := myjwt.ParseToken(mockJWT.String())
	if *actual != expected {
		t.Errorf(
			"TestParseToken(): expected %#v, actual %#v",
			expected,
			actual,
		)
	}
}

var parseJWTFailureTests = []struct {
	t        string
	expected error
}{
	// JWT expected to be headless with format "payload.signature"
	{"", myjwt.ErrInvalidJWTFormat},
	{"foo.bar.baz", myjwt.ErrInvalidJWTFormat},
	{"foo.bar", myjwt.ErrInvalidJWTEncoding},
	{mockJWTString, myjwt.ErrInvalidJWTPayload},
}

func TestParseTokenFailure(t *testing.T) {
	for _, tt := range parseJWTFailureTests {
		_, actual := myjwt.ParseToken(tt.t)
		if actual != tt.expected {
			t.Errorf(
				"TestParseToken(%s): expected %#v, actual %#v",
				tt.t,
				tt.expected,
				actual,
			)
		}
	}
}

package myaws_test

import (
	"testing"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/jwt"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
)

var testPayload = jwt.NewPayload(&jwt.NewPayloadInput{
	Iat: time.Now(),
	Exp: time.Now().Add(time.Minute * time.Duration(10)),
	Id:  "asdf",
})
var testToken = jwt.Token{Payload: testPayload, Signature: testPayload.String()}

var mockJwtKms = myaws.NewMockJwtKms()

func TestEncrypt(t *testing.T) {
	payload := testPayload
	expected := testToken
	actual := mockJwtKms.Encrypt(&payload)
	if *actual != expected {
		t.Errorf(
			"TestEncrypt(%+v): expected %s, actual %s",
			payload,
			expected,
			actual,
		)
	}
}

func TestEncryptPanic(t *testing.T) {
	myaws.MockKmsServiceError = true
	defer func() {
		myaws.MockKmsServiceError = false
	}()
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Expected panic from AWS error")
		}
	}()
	payload := testPayload
	mockJwtKms.Encrypt(&payload)
}

func TestDecryptSuccess(t *testing.T) {
	token := testToken
	expected := testPayload
	actual, err := mockJwtKms.Decrypt(&token)
	if err != nil {
		t.Fatalf("Fatal TestDecrypt(%s): %s", token, err)
	}
	if *actual != expected {
		t.Errorf(
			"TestDecrypt(%s): expected %#v, actual %#v",
			token,
			expected,
			actual,
		)
	}
}

var decryptFailureTests = []struct {
	t        jwt.Token
	expected error
}{
	{
		(jwt.Token{
			Payload: jwt.NewPayload(&jwt.NewPayloadInput{
				Iat: time.Now().Add(time.Minute * time.Duration(11)),
			}),
		}),
		myaws.ErrTokenIssuedAfter,
	},
	{
		(jwt.Token{
			Payload: jwt.NewPayload(&jwt.NewPayloadInput{
				Exp: time.Now().Add(-time.Minute * time.Duration(1)),
			}),
		}),
		myaws.ErrTokenExpired,
	},
	{
		(jwt.Token{Signature: "invalid"}),
		myaws.ErrInvalidSignature,
	},
}

func TestDecryptFailure(t *testing.T) {
	for _, tt := range decryptFailureTests {
		_, actual := mockJwtKms.Decrypt(&tt.t)
		if actual != tt.expected {
			t.Errorf(
				"TestDecrypt(%s): expected %#v, actual %#v",
				tt.t,
				tt.expected,
				actual,
			)
		}
	}
}

func TestDecryptPanic(t *testing.T) {
	myaws.MockKmsServiceError = true
	defer func() {
		myaws.MockKmsServiceError = false
	}()
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Expected panic from AWS error")
		}
	}()
	token := testToken
	mockJwtKms.Decrypt(&token)
}

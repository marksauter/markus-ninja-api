package data_test

import (
	"testing"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
)

var testPayload = myjwt.Payload{
	Iat: time.Now().Unix(),
	Exp: time.Now().Add(time.Minute * time.Duration(10)).Unix(),
	Sub: "asdf",
}
var testJWT = myjwt.JWT{Payload: testPayload, Signature: testPayload.String()}

var mockAuthService = data.NewMockAuthService()

func TestSignJWT(t *testing.T) {
	payload := testPayload
	expected := testJWT
	actual, _ := mockAuthService.SignJWT(&payload)
	if *actual != expected {
		t.Errorf(
			"TestSignJWT(%+v): expected %s, actual %s",
			payload,
			expected,
			actual,
		)
	}
}

func TestSignJWTServiceError(t *testing.T) {
	myaws.MockKMSServiceError = true
	defer func() {
		myaws.MockKMSServiceError = false
	}()
	payload := testPayload
	_, err := mockAuthService.SignJWT(&payload)
	if err == nil {
		t.Errorf("TestSignJWTServiceError(%#v): expected error from aws", payload)
	}
}

func TestValidateJWTSuccess(t *testing.T) {
	jwt := testJWT
	expected := testPayload
	actual, err := mockAuthService.ValidateJWT(&jwt)
	if err != nil {
		t.Fatalf("Fatal TestValidateJWT(%s): %s", jwt, err)
	}
	if *actual != expected {
		t.Errorf(
			"TestValidateJWT(%s): expected %#v, actual %#v",
			jwt,
			expected,
			actual,
		)
	}
}

var decryptFailureTests = []struct {
	t        myjwt.JWT
	expected error
}{
	{
		(myjwt.JWT{
			Payload: myjwt.Payload{
				Iat: time.Now().Add(time.Minute * time.Duration(11)).Unix(),
			},
		}),
		myjwt.ErrTokenIssuedAfter,
	},
	{
		(myjwt.JWT{
			Payload: myjwt.Payload{
				Exp: time.Now().Add(-time.Minute * time.Duration(1)).Unix(),
			},
		}),
		myjwt.ErrTokenExpired,
	},
	{
		(myjwt.JWT{
			Payload: myjwt.Payload{
				Exp: time.Now().Add(time.Minute * time.Duration(1)).Unix(),
			},
			Signature: "invalid",
		}),
		myjwt.ErrInvalidSignature,
	},
}

func TestValidateJWTFailure(t *testing.T) {
	for _, tt := range decryptFailureTests {
		_, actual := mockAuthService.ValidateJWT(&tt.t)
		if actual != tt.expected {
			t.Errorf(
				"TestValidateJWT(%s): expected %#v, actual %#v",
				tt.t,
				tt.expected,
				actual,
			)
		}
	}
}

func TestValidateJWTPanic(t *testing.T) {
	myaws.MockKMSServiceError = true
	defer func() {
		myaws.MockKMSServiceError = false
	}()
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Expected panic from AWS error")
		}
	}()
	jwt := testJWT
	mockAuthService.ValidateJWT(&jwt)
}

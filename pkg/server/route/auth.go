package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/justinas/alice"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/rs/cors"
)

type AuthErrorCode int

const (
	UnknownAuthError AuthErrorCode = iota
	InvalidRequest
	InvalidClient
	InvalidGrant
	InvalidScope
	UnauthorizedClient
	UnsupportedGrantType
)

func (e *AuthErrorCode) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*e = UnknownAuthError
	case "invalid_request":
		*e = InvalidRequest
	case "invalid_client":
		*e = InvalidClient
	case "invalid_grant":
		*e = InvalidGrant
	case "invalid_scope":
		*e = InvalidScope
	case "unauthorized_client":
		*e = UnauthorizedClient
	case "unsupported_grant_type":
		*e = UnsupportedGrantType
	}

	return nil
}

func (e AuthErrorCode) MarshalJSON() ([]byte, error) {
	var s string
	switch e {
	default:
		s = "unknown_auth_error"
	case InvalidRequest:
		s = "invalid_request"
	case InvalidClient:
		s = "invalid_client"
	case InvalidGrant:
		s = "invalid_grant"
	case InvalidScope:
		s = "invalid_scope"
	case UnauthorizedClient:
		s = "unauthorized_client"
	case UnsupportedGrantType:
		s = "unsupported_grant_type"
	}

	return json.Marshal(s)
}

type AuthTokenSuccessResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

func (r *AuthTokenSuccessResponse) WriteTo(rw http.ResponseWriter) error {
	json, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("AuthTokenSuccessResponse.WriteTo: %v", err)
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write(json)
	return nil
}

func (r *AuthTokenSuccessResponse) statusHTTP() int {
	return http.StatusOK
}

type AuthTokenErrorResponse struct {
	Error            AuthErrorCode `json:"error,omitempty"`
	ErrorDescription string        `json:"error_description,omitempty"`
	ErrorUri         string        `json:"error_uri,omitempty"`
}

func (r *AuthTokenErrorResponse) WriteTo(rw http.ResponseWriter) error {
	json, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("AuthTokenErrorResponse.WriteTo: %v", err)
	}
	rw.WriteHeader(r.statusHTTP())
	rw.Write(json)
	return nil
}

func (r *AuthTokenErrorResponse) statusHTTP() int {
	switch r.Error {
	case InvalidClient:
		return http.StatusUnauthorized
	default:
		return http.StatusBadRequest
	}
}

var Login = alice.New(LoginCors.Handler).Then(LoginHandler{})

var LoginCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"*"},
})

type LoginHandler struct {
	UserRepo *repo.UserRepo
}

func (h LoginHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	// ctx := req.Context()

	if req.Method != http.MethodPost {
		postMethodNotAllowed := myhttp.ErrorResponse{
			Error:            myhttp.MethodNotAllowed,
			ErrorDescription: fmt.Sprintf("route: %v method not allowed", req.Method),
		}
		postMethodNotAllowed.WriteTo(rw)
		return
	}
	err := req.ParseForm()
	if err != nil {
		parseFormFailed := myhttp.ErrorResponse{
			Error:            myhttp.InternalServerError,
			ErrorDescription: fmt.Sprintf("route: failed to parse form, %v", err),
		}
		parseFormFailed.WriteTo(rw)
		return
	}
	userCredentials := model.UserCredentials{
		Login:    req.PostFormValue("login"),
		Password: req.PostFormValue("password"),
	}
	// err = myhttp.UnmarshalRequestBody(req, &userCredentials)
	// if err != nil {
	//   requestBodyInvalid := AuthTokenErrorResponse{
	//     Error:            InvalidRequest,
	//     ErrorDescription: fmt.Sprintf("route: invalid request body, %v", err),
	//   }
	//   requestBodyInvalid.WriteTo(rw)
	//   return
	// }

	_, err = h.UserRepo.VerifyCredentials(&userCredentials)
	if err != nil {
		verificationFailed := AuthTokenErrorResponse{
			Error:            InvalidScope,
			ErrorDescription: fmt.Sprintf("route: verification failed, %v", err),
		}

		verificationFailed.WriteTo(rw)
		return
	}

	response := AuthTokenSuccessResponse{
		AccessToken: "token",
	}
	response.WriteTo(rw)
}

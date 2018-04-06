package route

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	InvalidRequest       ErrorCode = "invalid_request"
	InvalidClient        ErrorCode = "invalid_client"
	InvalidGrant         ErrorCode = "invalid_grant"
	InvalidScope         ErrorCode = "invalid_scope"
	UnauthorizedClient   ErrorCode = "unauthorized_client"
	UnsupportedGrantType ErrorCode = "unsupported_grant_type"
)

type Response interface {
	WriteTo(http.ResponseWriter)
}

type SuccessResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`

	Code int
}

func (r *SuccessResponse) WriteTo(rw http.ResponseWriter) error {
	json, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("SuccessResponse.WriteTo: %v", err)
	}
	rw.WriteHeader(r.Code)
	rw.Write(r)
	return nil
}

type ErrorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorUri         string `json:"error_uri,omitempty"`

	Code int
}

func (r *ErrorResponse) WriteTo(rw http.ResponseWriter) error {
	json, err := json.Marshall(r)
	if err != nil {
		return fmt.Errorf("ErrorResponse.WriteTo: %v", err)
	}
	rw.WriteHeader(r.Code)
	rw.Write(r)
	return nil
}

type LoginHandler struct{}

func (h LoginHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	ctx := req.Context()
	code := http.StatusOK

	if req.Method != http.MethodPost {
		code = http.StatusMethodNotAllowed
	}
}

package myhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Response interface {
	WriteTo(http.ResponseWriter) error
	StatusHTTP() int
}

type ErrorCode int

const (
	UnknownError ErrorCode = iota
	InternalServerError
	MethodNotAllowed
)

func (e *ErrorCode) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*e = UnknownError
	case "internal_server_error":
		*e = InternalServerError
	case "method_not_allowed":
		*e = MethodNotAllowed
	}

	return nil
}

func (e ErrorCode) MarshalJSON() ([]byte, error) {
	var s string
	switch e {
	default:
		s = "unknown_error"
	case InternalServerError:
		s = "internal_server_error"
	case MethodNotAllowed:
		s = "method_not_allowed"
	}

	return json.Marshal(s)
}

type ErrorResponse struct {
	Error            ErrorCode `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ErrorUri         string    `json:"error_uri,omitempty"`
}

func (r *ErrorResponse) WriteTo(rw http.ResponseWriter) error {
	json, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("ErrorResponse.WriteTo: %v", err)
	}
	rw.WriteHeader(r.statusHTTP())
	rw.Write(json)
	return nil
}

func (r *ErrorResponse) statusHTTP() int {
	switch r.Error {
	default:
		return http.StatusBadRequest
	case MethodNotAllowed:
		return http.StatusMethodNotAllowed
	}
}

func UnmarshalRequestBody(req *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("myhttp: %v", err)
	}
	return json.Unmarshal(body, v)
}

package myhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ErrorCode int

const (
	UnknownError ErrorCode = iota
	AccessDenied
	InvalidRequest
	InternalServerError
	InvalidPassword
	InvalidCredentials
	MethodNotAllowed
	PasswordStrengthError
	TooManyAttempts
	Unauthorized
	UserExists
	UsernameExists
)

func (e *ErrorCode) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*e = UnknownError
	case "access_denied":
		*e = AccessDenied
	case "invalid_credentials":
		*e = InvalidCredentials
	case "invalid_request":
		*e = InvalidRequest
	case "internal_server_error":
		*e = InternalServerError
	case "invalid_password":
		*e = InvalidPassword
	case "method_not_allowed":
		*e = MethodNotAllowed
	case "password_strength_error":
		*e = PasswordStrengthError
	case "too_many_attempts":
		*e = TooManyAttempts
	case "unauthorized":
		*e = Unauthorized
	case "user_exists":
		*e = UserExists
	case "username_exists":
		*e = UsernameExists
	}

	return nil
}

func (e ErrorCode) MarshalJSON() ([]byte, error) {
	var s string
	switch e {
	default:
		s = "unknown_error"
	case AccessDenied:
		s = "access_denied"
	case InvalidCredentials:
		s = "invalid_credentials"
	case InvalidRequest:
		s = "invalid_request"
	case InternalServerError:
		s = "internal_server_error"
	case InvalidPassword:
		s = "invalid_password"
	case MethodNotAllowed:
		s = "method_not_allowed"
	case PasswordStrengthError:
		s = "password_strength_error"
	case TooManyAttempts:
		s = "too_many_attempts"
	case Unauthorized:
		s = "unauthorized"
	case UserExists:
		s = "user_exists"
	case UsernameExists:
		s = "username_exists"
	}

	return json.Marshal(s)
}

type ErrorResponse struct {
	Error            ErrorCode `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ErrorUri         string    `json:"error_uri,omitempty"`
}

func (r *ErrorResponse) StatusHTTP() int {
	switch r.Error {
	default:
		return http.StatusBadRequest
	case AccessDenied:
		return http.StatusForbidden
	case InvalidCredentials:
		return http.StatusUnprocessableEntity
	case InvalidPassword:
		return http.StatusUnprocessableEntity
	case InvalidRequest:
		return http.StatusBadRequest
	case InternalServerError:
		return http.StatusInternalServerError
	case MethodNotAllowed:
		return http.StatusMethodNotAllowed
	case PasswordStrengthError:
		return http.StatusUnprocessableEntity
	case TooManyAttempts:
		return http.StatusTooManyRequests
	case Unauthorized:
		return http.StatusUnauthorized
	}
}

func AccessDeniedErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Error:            AccessDenied,
		ErrorDescription: "You do not have the necessary permissions to access this content",
	}
}

func InvalidRequestErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{
		Error:            InvalidRequest,
		ErrorDescription: err,
	}
}

func InternalServerErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{
		Error:            InternalServerError,
		ErrorDescription: fmt.Sprintf("Something went wrong: %v", err),
	}
}

func InvalidPasswordResponse() *ErrorResponse {
	return &ErrorResponse{
		Error:            InvalidPassword,
		ErrorDescription: "The password doesn't comply with the password policy for the connection",
	}
}

func InvalidCredentialsErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Error:            InvalidCredentials,
		ErrorDescription: "The credentials used for authentication are invalid",
	}
}

func MethodNotAllowedResponse(method string) *ErrorResponse {
	method = strings.ToUpper(method)
	return &ErrorResponse{
		Error:            MethodNotAllowed,
		ErrorDescription: fmt.Sprintf("%v method not allowed", method),
	}
}

func PasswordStrengthErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{
		Error:            PasswordStrengthError,
		ErrorDescription: err,
	}
}

func TooManyAttemptsResponse() *ErrorResponse {
	return &ErrorResponse{
		Error:            TooManyAttempts,
		ErrorDescription: "The account is blocked due to too many attempts to sign in",
	}
}

func UnauthorizedErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{
		Error:            Unauthorized,
		ErrorDescription: err,
	}
}

func UserExistsResponse() *ErrorResponse {
	return &ErrorResponse{
		Error:            UserExists,
		ErrorDescription: "The user you are attempting to sign up has already signed up",
	}
}

func UsernameExistsResponse() *ErrorResponse {
	return &ErrorResponse{
		Error:            UsernameExists,
		ErrorDescription: "The username you are attempting to sign up with is already in use",
	}
}

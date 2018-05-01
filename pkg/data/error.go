package data

import (
	"fmt"
	"strings"
)

type DataFieldErrorCode int

const (
	Unknown DataFieldErrorCode = iota
	DuplicateField
	RequiredField
)

func (c DataFieldErrorCode) String() string {
	switch c {
	default:
		return "unknown"
	case DuplicateField:
		return "duplicate_field"
	case RequiredField:
		return "required_field"
	}
}

type DataFieldError struct {
	Code  DataFieldErrorCode
	Field string
}

func (e DataFieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Field)
}

func ParseConstraintName(constraintName string) (field string) {
	parsedContraintName := strings.Split(constraintName, "_")
	field = strings.Join(
		parsedContraintName[1:len(parsedContraintName)-1],
		"_",
	)
	return
}

func DuplicateFieldError(field string) DataFieldError {
	return DataFieldError{DuplicateField, field}
}

func RequiredFieldError(field string) DataFieldError {
	return DataFieldError{RequiredField, field}
}

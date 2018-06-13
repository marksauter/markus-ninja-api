package data

import (
	"fmt"
	"strings"
)

type DataFieldErrorCode int

const (
	UnknownDataFieldErrorCode DataFieldErrorCode = iota
	DuplicateField
	RequiredField
)

func (c DataFieldErrorCode) String() string {
	switch c {
	default:
		return "unknown_data_field_error"
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
	parsedContraintName := strings.Split(constraintName, "__")
	if len(parsedContraintName) > 1 {
		field = strings.Join(
			parsedContraintName[1:len(parsedContraintName)-1],
			"_",
		)
	} else {
		field = constraintName
	}
	return
}

func DuplicateFieldError(field string) DataFieldError {
	return DataFieldError{DuplicateField, field}
}

func RequiredFieldError(field string) DataFieldError {
	return DataFieldError{RequiredField, field}
}

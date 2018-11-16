package myerr

import (
	"errors"
	"fmt"
)

var SomethingWentWrongError = errors.New("Something went wrong")

type UnexpectedError struct {
	Message string
}

func (e UnexpectedError) Error() string {
	return fmt.Sprintf("UNEXPECTED: %s", e.Message)
}

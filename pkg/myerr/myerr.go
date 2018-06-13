package myerr

import "fmt"

type UnexpectedError struct {
	Message string
}

func (e UnexpectedError) Error() string {
	return fmt.Sprintf("UNEXPECTED: %s", e.Message)
}

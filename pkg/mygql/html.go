package mygql

import (
	"bytes"
	"fmt"

	"golang.org/x/net/html"
)

type HTML string

func (h HTML) ImplementsGraphQLType(name string) bool {
	return name == "HTML"
}

func (h HTML) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case html.Node:
		h = HTML(input.Data)
		return nil
	case string:
		var err error
		h = HTML(input)
		_, err = html.Parse(bytes.NewBufferString(input))
		return err
	default:
		return fmt.Errorf("HTML cannot represent an invalid HTML string")
	}
}

package mygql

import (
	"fmt"
	"net/url"
)

type URI string

func (u URI) ImplementsGraphQLType(name string) bool {
	return name == "URI"
}

func (u URI) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case url.URL:
		u = URI(input.String())
		return nil
	case string:
		var err error
		u = URI(input)
		_, err = url.Parse(input)
		return err
	default:
		return fmt.Errorf("URI cannot represent an invalid URI string")
	}
}

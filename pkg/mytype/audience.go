package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type Audience int

const (
	NoAudience Audience = iota
	Authenticated
	Everyone
)

func (a Audience) String() string {
	switch a {
	case Authenticated:
		return "AUTHENTICATED"
	case Everyone:
		return "EVERYONE"
	default:
		return "NOAUDIENCE"
	}
}

func ParseAudience(aud string) (Audience, error) {
	switch strings.ToUpper(aud) {
	case "AUTHENTICATED":
		return Authenticated, nil
	case "EVERYONE":
		return Everyone, nil
	default:
		var a Audience
		return a, fmt.Errorf("invalid audience: %q", aud)
	}
}

func (a *Audience) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case string:
		*a, err = ParseAudience(v)
		return
	default:
		return fmt.Errorf("invalid type for audience %T", v)
	}
}

func (a Audience) Value() (driver.Value, error) {
	return a.String(), nil
}

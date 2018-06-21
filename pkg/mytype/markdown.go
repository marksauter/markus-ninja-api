package mytype

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type Markdown struct {
	Status pgtype.Status
	String string
}

var atRef = regexp.MustCompile(`(?:|\s+)@(\w+)(?:\s+|$)`)

func (src *Markdown) AtRefs() []string {
	result := atRef.FindAllStringSubmatch(src.String, -1)
	refs := make([]string, 0, len(result))
	for _, r := range result {
		if r[1] != "" {
			// The group '(\w+)' match will be at position 1 in 'r'
			refs = append(refs, r[1])
		}
	}
	return refs
}

var numberRef = regexp.MustCompile(`(?:^|\s+)#(\d)(?:\s+|$)`)

func (src *Markdown) NumberRefs() ([]int32, error) {
	result := numberRef.FindAllStringSubmatch(src.String, -1)
	refs := make([]int32, 0, len(result))
	for _, r := range result {
		if r[1] != "" {
			// The group '(\d)' match will be at position 1 in 'r'
			n, err := strconv.ParseInt(r[1], 10, 32)
			if err != nil {
				return nil, err
			}
			refs = append(refs, int32(n))
		}
	}
	return refs, nil
}

type CrossStudyRef struct {
	Owner  string
	Name   string
	Number int32
}

var crossStudyRef = regexp.MustCompile(`(?:^|\s+)(\w+)/([\w|-]+)#(\d)(?:\s+|$)`)

func (src *Markdown) CrossStudyRefs() ([]*CrossStudyRef, error) {
	result := crossStudyRef.FindAllStringSubmatch(src.String, -1)
	refs := make([]*CrossStudyRef, 0, len(result))
	for _, r := range result {
		owner := r[1]
		name := r[2]
		number, err := strconv.ParseInt(r[3], 10, 32)
		if err != nil {
			return nil, err
		}
		ref := &CrossStudyRef{
			Owner:  owner,
			Name:   name,
			Number: int32(number),
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func (src *Markdown) ToHTML() string {
	return string(util.MarkdownToHTML([]byte(src.String)))
}

func (src *Markdown) ToText() string {
	return util.MarkdownToText(src.String)
}

func (dst *Markdown) Set(src interface{}) error {
	if src == nil {
		*dst = Markdown{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case Markdown:
		*dst = value
	case *Markdown:
		*dst = *value
	case string:
		*dst = Markdown{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = Markdown{Status: pgtype.Null}
		} else {
			*dst = Markdown{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = Markdown{Status: pgtype.Null}
		} else {
			*dst = Markdown{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to Markdown", value)
	}

	return nil
}

func (dst *Markdown) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Markdown) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *string:
			*v = src.String
			return nil
		case *[]byte:
			*v = make([]byte, len(src.String))
			copy(*v, src.String)
			return nil
		default:
			if nextDst, retry := pgtype.GetAssignToDstType(dst); retry {
				return src.AssignTo(nextDst)
			}
		}
	case pgtype.Null:
		return pgtype.NullAssignTo(dst)
	}

	return fmt.Errorf("cannot decode %v into %T", src, dst)
}

func (dst *Markdown) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Markdown{Status: pgtype.Null}
		return nil
	}

	*dst = Markdown{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *Markdown) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *Markdown) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *Markdown) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Markdown) Scan(src interface{}) error {
	if src == nil {
		*dst = Markdown{Status: pgtype.Null}
		return nil
	}

	switch src := src.(type) {
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		return dst.DecodeText(nil, srcCopy)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src *Markdown) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}

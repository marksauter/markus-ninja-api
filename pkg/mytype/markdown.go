package mytype

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

// Markdown -
type Markdown struct {
	Status pgtype.Status
	String string
}

var AtRefRegexp = regexp.MustCompile(`(?:^|\s)@(\w+)(?:\s|$)`)

// AtRef -
type AtRef struct {
	Name  string
	Index []int
}

// AtRefs -
func (src *Markdown) AtRefs() []*AtRef {
	result := AtRefRegexp.FindAllStringSubmatchIndex(src.String, -1)
	refs := make([]*AtRef, len(result))
	for i, r := range result {
		name := src.String[r[2]:r[3]]
		if name != "" {
			ref := &AtRef{
				Name:  name,
				Index: []int{r[0], r[1]},
			}
			refs[i] = ref
		}
	}
	return refs
}

var AssetRefRegexp = regexp.MustCompile(`(?:(?:^|\s|\[)\${2})([\w-.]+)(?:\]|\s|$)`)

// AssetRef -
type AssetRef struct {
	Name  string
	Index []int
}

// AssetRefs -
func (src *Markdown) AssetRefs() []*AssetRef {
	result := AssetRefRegexp.FindAllStringSubmatchIndex(src.String, -1)
	refs := make([]*AssetRef, len(result))
	for i, r := range result {
		name := src.String[r[2]:r[3]]
		if name != "" {
			ref := &AssetRef{
				Name:  name,
				Index: []int{r[0], r[1]},
			}
			refs[i] = ref
		}
	}
	return refs
}

var NumberRefRegexp = regexp.MustCompile(`(?:^|\s)#(\d+)(?:\s|$)`)

// NumberRef -
type NumberRef struct {
	Number int32
	Index  []int
}

// NumberRefs -
func (src *Markdown) NumberRefs() ([]*NumberRef, error) {
	result := NumberRefRegexp.FindAllStringSubmatchIndex(src.String, -1)
	refs := make([]*NumberRef, len(result))
	for i, r := range result {
		number := src.String[r[2]:r[3]]
		if number != "" {
			n, err := strconv.ParseInt(number, 10, 32)
			if err != nil {
				return nil, err
			}
			ref := &NumberRef{
				Number: int32(n),
				Index:  []int{r[0], r[1]},
			}
			refs[i] = ref
		}
	}
	return refs, nil
}

// CrossStudyRef -
type CrossStudyRef struct {
	Owner  string
	Name   string
	Number int32
	Index  []int
}

var CrossStudyRefRegexp = regexp.MustCompile(`(?:^|\s)(\w+)\/([\w-]+)#(\d+)(?:\s|$)`)

// CrossStudyRefs -
func (src *Markdown) CrossStudyRefs() ([]*CrossStudyRef, error) {
	result := CrossStudyRefRegexp.FindAllStringSubmatchIndex(src.String, -1)
	refs := make([]*CrossStudyRef, len(result))
	for i, r := range result {
		owner := src.String[r[2]:r[3]]
		name := src.String[r[4]:r[5]]
		number, err := strconv.ParseInt(src.String[r[6]:r[7]], 10, 32)
		if err != nil {
			return nil, err
		}
		ref := &CrossStudyRef{
			Owner:  owner,
			Name:   name,
			Number: int32(number),
			Index:  []int{r[0], r[1]},
		}
		refs[i] = ref
	}
	return refs, nil
}

// ToHTML -
func (src *Markdown) ToHTML(clientURL, studyResourcePath string) (string, error) {
	src.String = AtRefRegexp.ReplaceAllString(src.String, "[@$1]("+clientURL+"/$1)")
	src.String = NumberRefRegexp.ReplaceAllString(src.String, "[#$1]("+clientURL+studyResourcePath+"/lesson/$1)")
	src.String = CrossStudyRefRegexp.ReplaceAllString(src.String, "[$1/$2#$3]("+clientURL+"/$1/$2"+"/lesson/$3)")
	return string(util.MarkdownToHTML([]byte(src.String))), nil
}

// ToText -
func (src *Markdown) ToText() string {
	return util.MarkdownToText(src.String)
}

// Set -
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

// Get -
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

// AssignTo -
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

// DecodeText -
func (dst *Markdown) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Markdown{Status: pgtype.Null}
		return nil
	}

	*dst = Markdown{String: string(src), Status: pgtype.Present}
	return nil
}

// DecodeBinary -
func (dst *Markdown) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

// EncodeText -
func (src *Markdown) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

// EncodeBinary -
func (src *Markdown) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan - implements the database/sql Scanner interface.
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

// Value - implements the database/sql/driver Valuer interface.
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

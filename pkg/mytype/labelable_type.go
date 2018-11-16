package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type LabelableTypeValue int

const (
	LabelableTypeLesson LabelableTypeValue = iota
	LabelableTypeLessonComment
)

func (f LabelableTypeValue) String() string {
	switch f {
	case LabelableTypeLesson:
		return "Lesson"
	case LabelableTypeLessonComment:
		return "LessonComment"
	default:
		return "unknown"
	}
}

type LabelableType struct {
	Status pgtype.Status
	V      LabelableTypeValue
}

func NewLabelableType(v LabelableTypeValue) LabelableType {
	return LabelableType{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseLabelableType(s string) (LabelableType, error) {
	switch strings.Title(s) {
	case "Lesson":
		return LabelableType{
			Status: pgtype.Present,
			V:      LabelableTypeLesson,
		}, nil
	case "LessonComment":
		return LabelableType{
			Status: pgtype.Present,
			V:      LabelableTypeLessonComment,
		}, nil
	default:
		var f LabelableType
		return f, fmt.Errorf("invalid LabelableType: %q", s)
	}
}

func (src *LabelableType) String() string {
	return src.V.String()
}

func (dst *LabelableType) Set(src interface{}) error {
	if src == nil {
		*dst = LabelableType{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case LabelableType:
		*dst = value
		dst.Status = pgtype.Present
	case *LabelableType:
		*dst = *value
		dst.Status = pgtype.Present
	case LabelableTypeValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *LabelableTypeValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseLabelableType(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseLabelableType(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseLabelableType(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to LabelableType", value)
	}

	return nil
}

func (src *LabelableType) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *LabelableType) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *string:
			*v = src.V.String()
			return nil
		case *[]byte:
			*v = make([]byte, len(src.V.String()))
			copy(*v, src.V.String())
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

func (dst *LabelableType) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = LabelableType{Status: pgtype.Null}
		return nil
	}

	t, err := ParseLabelableType(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *LabelableType) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *LabelableType) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *LabelableType) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *LabelableType) Scan(src interface{}) error {
	if src == nil {
		*dst = LabelableType{Status: pgtype.Null}
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
func (src *LabelableType) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}

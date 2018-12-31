package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type CommentableTypeValue int

const (
	CommentableTypeLesson CommentableTypeValue = iota
	CommentableTypeUserAsset
)

func (f CommentableTypeValue) String() string {
	switch f {
	case CommentableTypeLesson:
		return "Lesson"
	case CommentableTypeUserAsset:
		return "UserAsset"
	default:
		return "unknown"
	}
}

type CommentableType struct {
	Status pgtype.Status
	V      CommentableTypeValue
}

func NewCommentableType(v CommentableTypeValue) CommentableType {
	return CommentableType{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseCommentableType(s string) (CommentableType, error) {
	switch strings.Title(s) {
	case "Lesson":
		return CommentableType{
			Status: pgtype.Present,
			V:      CommentableTypeLesson,
		}, nil
	case "UserAsset":
		return CommentableType{
			Status: pgtype.Present,
			V:      CommentableTypeUserAsset,
		}, nil
	default:
		var f CommentableType
		return f, fmt.Errorf("invalid CommentableType: %q", s)
	}
}

func (src *CommentableType) String() string {
	return src.V.String()
}

func (dst *CommentableType) Set(src interface{}) error {
	if src == nil {
		*dst = CommentableType{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case CommentableType:
		*dst = value
		dst.Status = pgtype.Present
	case *CommentableType:
		*dst = *value
		dst.Status = pgtype.Present
	case CommentableTypeValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *CommentableTypeValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseCommentableType(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseCommentableType(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseCommentableType(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to CommentableType", value)
	}

	return nil
}

func (src *CommentableType) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *CommentableType) AssignTo(dst interface{}) error {
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

func (dst *CommentableType) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = CommentableType{Status: pgtype.Null}
		return nil
	}

	t, err := ParseCommentableType(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *CommentableType) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *CommentableType) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *CommentableType) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *CommentableType) Scan(src interface{}) error {
	if src == nil {
		*dst = CommentableType{Status: pgtype.Null}
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
func (src *CommentableType) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}

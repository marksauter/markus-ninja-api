package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type TopicableTypeValue int

const (
	TopicableTypeCourse TopicableTypeValue = iota
	TopicableTypeStudy
)

func (f TopicableTypeValue) String() string {
	switch f {
	case TopicableTypeCourse:
		return "Course"
	case TopicableTypeStudy:
		return "Study"
	default:
		return "unknown"
	}
}

type TopicableType struct {
	Status pgtype.Status
	V      TopicableTypeValue
}

func NewTopicableType(v TopicableTypeValue) TopicableType {
	return TopicableType{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseTopicableType(s string) (TopicableType, error) {
	switch strings.Title(s) {
	case "Course":
		return TopicableType{
			Status: pgtype.Present,
			V:      TopicableTypeCourse,
		}, nil
	case "Study":
		return TopicableType{
			Status: pgtype.Present,
			V:      TopicableTypeStudy,
		}, nil
	default:
		var f TopicableType
		return f, fmt.Errorf("invalid TopicableType: %q", s)
	}
}

func (src *TopicableType) String() string {
	return src.V.String()
}

func (dst *TopicableType) Set(src interface{}) error {
	if src == nil {
		*dst = TopicableType{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case TopicableType:
		*dst = value
		dst.Status = pgtype.Present
	case *TopicableType:
		*dst = *value
		dst.Status = pgtype.Present
	case TopicableTypeValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *TopicableTypeValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseTopicableType(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseTopicableType(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseTopicableType(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to TopicableType", value)
	}

	return nil
}

func (src *TopicableType) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *TopicableType) AssignTo(dst interface{}) error {
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

func (dst *TopicableType) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = TopicableType{Status: pgtype.Null}
		return nil
	}

	t, err := ParseTopicableType(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *TopicableType) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *TopicableType) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *TopicableType) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *TopicableType) Scan(src interface{}) error {
	if src == nil {
		*dst = TopicableType{Status: pgtype.Null}
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
func (src *TopicableType) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}

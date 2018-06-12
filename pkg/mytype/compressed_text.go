package mytype

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type Body struct {
	Status pgtype.Status
	String string
}

func (dst *Body) compress() (err error) {
	dst.String, err = util.CompressString(dst.String)
	return
}

func (dst *Body) decompress() (err error) {
	dst.String, err = util.DecompressString(dst.String)
	return
}

func (dst *Body) Set(src interface{}) error {
	if src == nil {
		*dst = Body{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case Body:
		*dst = value
	case *Body:
		*dst = *value
	case string:
		*dst = Body{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = Body{Status: pgtype.Null}
		} else {
			*dst = Body{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = Body{Status: pgtype.Null}
		} else {
			*dst = Body{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to Body", value)
	}

	return nil
}

func (dst *Body) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Body) AssignTo(dst interface{}) error {
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

func (dst *Body) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Body{Status: pgtype.Null}
		return nil
	}

	*dst = Body{String: string(src), Status: pgtype.Present}
	if err := dst.decompress(); err != nil {
		return err
	}
	return nil
}

func (dst *Body) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *Body) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	if err := src.compress(); err != nil {
		return nil, err
	}
	return append(buf, src.String...), nil
}

func (src *Body) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Body) Scan(src interface{}) error {
	if src == nil {
		*dst = Body{Status: pgtype.Null}
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
func (src *Body) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		if err := src.compress(); err != nil {
			return nil, err
		}
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}

package mytype

import (
	"database/sql/driver"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/pgtype"
)

type URLSafeName struct {
	Status pgtype.Status
	String string
}

var invalidURLSafeNameChar = regexp.MustCompile(`\W+`)

const urlSafeNameReplacementStr = "-"

func (dst *URLSafeName) Set(src interface{}) error {
	if src == nil {
		*dst = URLSafeName{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case URLSafeName:
		*dst = value
	case *URLSafeName:
		*dst = *value
	case string:
		value = invalidURLSafeNameChar.ReplaceAllString(value, urlSafeNameReplacementStr)
		*dst = URLSafeName{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = URLSafeName{Status: pgtype.Null}
		} else {
			*value = invalidURLSafeNameChar.ReplaceAllString(*value, urlSafeNameReplacementStr)
			*dst = URLSafeName{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = URLSafeName{Status: pgtype.Null}
		} else {
			value = invalidURLSafeNameChar.ReplaceAll(value, []byte(urlSafeNameReplacementStr))
			*dst = URLSafeName{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to URLSafeName", value)
	}

	return nil
}

func (dst *URLSafeName) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *URLSafeName) AssignTo(dst interface{}) error {
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

func (dst *URLSafeName) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = URLSafeName{Status: pgtype.Null}
		return nil
	}

	*dst = URLSafeName{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *URLSafeName) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *URLSafeName) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *URLSafeName) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *URLSafeName) Scan(src interface{}) error {
	if src == nil {
		*dst = URLSafeName{Status: pgtype.Null}
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
func (src *URLSafeName) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}

package mytype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type Operation struct {
	// Access level for the Operation.
	AccessLevel AccessLevel
	// Type of node for the Operation.
	NodeType NodeType
	Status   pgtype.Status
}

func (o *Operation) String() string {
	return o.AccessLevel.String() + " " + o.NodeType.String()
}

func NewOperation(al AccessLevel, nt NodeType) *Operation {
	return &Operation{AccessLevel: al, NodeType: nt, Status: pgtype.Present}
}

var errInvalidOperation = errors.New("invalid Operation")

func ParseOperation(operation string) (*Operation, error) {
	o := &Operation{}

	parsedOperation := strings.SplitN(operation, " ", 2)
	if len(parsedOperation) != 2 {
		return o, fmt.Errorf("invalid operation: %q", operation)
	}
	accessLevel, err := ParseAccessLevel(parsedOperation[0])
	if err != nil {
		return o, fmt.Errorf("invalid operation: %v", err)
	}
	o.AccessLevel = accessLevel
	nodeType, err := ParseNodeType(parsedOperation[1])
	if err != nil {
		return o, fmt.Errorf("invalid operation: %v", err)
	}
	o.NodeType = nodeType

	return o, nil
}

func (dst *Operation) UnmarshalJSON(bs []byte) error {
	var s string
	if err := json.Unmarshal(bs, &s); err != nil {
		return err
	}
	o, err := ParseOperation(s)
	if err != nil {
		return err
	}
	*dst = *o
	return nil
}

func (dst *Operation) Set(src interface{}) error {
	if src == nil {
		*dst = Operation{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case Operation:
		*dst = value
	case *Operation:
		*dst = *value
	case string:
		o, err := ParseOperation(value)
		if err != nil {
			return err
		}
		*dst = *o
	case *string:
		if value == nil {
			*dst = Operation{Status: pgtype.Null}
		} else {
			o, err := ParseOperation(*value)
			if err != nil {
				return err
			}
			*dst = *o
		}
	case []byte:
		if value == nil {
			*dst = Operation{Status: pgtype.Null}
		} else {
			o, err := ParseOperation(string(value))
			if err != nil {
				return err
			}
			*dst = *o
		}
	default:
		return fmt.Errorf("cannot convert %v to Operation", value)
	}

	return nil
}

func (dst *Operation) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String()
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Operation) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *string:
			*v = src.String()
			return nil
		case *[]byte:
			*v = make([]byte, len(src.String()))
			copy(*v, src.String())
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

func (dst *Operation) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Operation{Status: pgtype.Null}
		return nil
	}

	o, err := ParseOperation(string(src))
	if err != nil {
		return err
	}
	*dst = *o
	return nil
}

func (dst *Operation) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *Operation) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String()...), nil
}

func (src *Operation) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Operation) Scan(src interface{}) error {
	if src == nil {
		*dst = Operation{Status: pgtype.Null}
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
func (src *Operation) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}

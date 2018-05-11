package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

// type LessonOrderField struct {
//   name  string
//   value data.OrderFieldValue
// }
//
// func ParseLessonOrderField(s string) (*LessonOrderField, error) {
//   f := &LessonOrderField{}
//   switch strings.ToLower(s) {
//   case "number":
//     f.name = "number"
//     f.value = &pgtype.Int4{Int: 0, Status: pgtype.Present}
//     return f, nil
//   default:
//     return nil, fmt.Errorf("invalid LessonOrderField: %q", s)
//   }
// }
//
// func (f *LessonOrderField) DecodeCursor(cursor string) error {
//   bs, err := base64.StdEncoding.DecodeString(cursor)
//   if err != nil {
//     return err
//   }
//   s := strings.TrimPrefix(string(bs), "cursor")
//   switch strings.ToLower(f.name) {
//   case "number":
//     value, err := strconv.ParseInt(s, 10, 32)
//     if err != nil {
//       return err
//     }
//     f.value.Set(value)
//     return nil
//   default:
//     return fmt.Errorf("invalid LessonOrderField: %q", f.name)
//   }
// }
//
// func (f *LessonOrderField) EncodeCursor(src *repo.LessonPermit) (string, error) {
//   switch strings.ToLower(f.name) {
//   case "number":
//     number, ok := src.Number()
//     if !ok {
//       return "", fmt.Errorf("invalid type %t for field %s", src, f.name)
//     }
//     cursor := base64.StdEncoding.EncodeToString(
//       []byte(fmt.Sprintf("cursor%d", number)),
//     )
//     return cursor, nil
//   default:
//     return "", fmt.Errorf("invalid LessonOrderField: %q", f.name)
//   }
// }
//
// func (f *LessonOrderField) Name() string {
//   return f.name
// }
//
// func (f *LessonOrderField) Value() OrderFieldValue {
//   return f.value
// }

type LessonOrderField int

const (
	LessonNumber LessonOrderField = iota
)

func ParseLessonOrderField(s string) (LessonOrderField, error) {
	switch strings.ToLower(s) {
	case "number":
		return LessonNumber, nil
	default:
		var f LessonOrderField
		return f, fmt.Errorf("invalid LessonOrderField: %q", s)
	}
}

func (f LessonOrderField) String() string {
	switch f {
	case LessonNumber:
		return "number"
	default:
		return "unknown"
	}
}

type LessonOrder struct {
	direction data.OrderDirection
	field     LessonOrderField
}

func (o *LessonOrder) Direction() string {
	return o.direction.String()
}

func (o *LessonOrder) Field() string {
	return o.field.String()
}

type LessonOrderArg struct {
	Direction string
	Field     string
}

func ParseLessonOrder(arg *LessonOrderArg) (*LessonOrder, error) {
	if arg == nil {
		arg = &LessonOrderArg{
			Direction: "ASC",
			Field:     "number",
		}
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseLessonOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	lessonOrder := &LessonOrder{
		direction: direction,
		field:     field,
	}
	return lessonOrder, nil
}

type lessonOrderResolver struct {
	LessonOrder
}

func (r *lessonOrderResolver) Direction() string {
	return r.LessonOrder.Direction()
}

func (r *lessonOrderResolver) Field() string {
	return r.LessonOrder.Field()
}

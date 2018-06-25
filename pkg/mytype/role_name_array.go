package mytype

// import (
//   "database/sql/driver"
//   "encoding/binary"
//   "fmt"
//
//   "github.com/jackc/pgx/pgio"
//   "github.com/jackc/pgx/pgtype"
// )
//
// type RoleNameArray struct {
//   Dimensions []pgtype.ArrayDimension
//   Elements   []RoleName
//   Status     pgtype.Status
// }
//
// func (dst *RoleNameArray) Set(src interface{}) error {
//   // untyped nil and typed nil interfaces are different
//   if src == nil {
//     *dst = RoleNameArray{Status: pgtype.Null}
//     return nil
//   }
//
//   switch value := src.(type) {
//
//   case []string:
//     if value == nil {
//       *dst = RoleNameArray{Status: pgtype.Null}
//     } else if len(value) == 0 {
//       *dst = RoleNameArray{Status: pgtype.Present}
//     } else {
//       elements := make([]RoleName, len(value))
//       for i := range value {
//         if err := elements[i].Set(value[i]); err != nil {
//           return err
//         }
//       }
//       *dst = RoleNameArray{
//         Elements:   elements,
//         Dimensions: []pgtype.ArrayDimension{{Length: int32(len(elements)), LowerBound: 1}},
//         Status:     pgtype.Present,
//       }
//     }
//   case []RoleNameValue:
//     if value == nil {
//       *dst = RoleNameArray{Status: pgtype.Null}
//     } else if len(value) == 0 {
//       *dst = RoleNameArray{Status: pgtype.Present}
//     } else {
//       elements := make([]RoleName, len(value))
//       for i := range value {
//         if err := elements[i].Set(value[i]); err != nil {
//           return err
//         }
//       }
//       *dst = RoleNameArray{
//         Elements:   elements,
//         Dimensions: []pgtype.ArrayDimension{{Length: int32(len(elements)), LowerBound: 1}},
//         Status:     pgtype.Present,
//       }
//     }
//
//   default:
//     return fmt.Errorf("cannot convert %v to RoleNameArray", value)
//   }
//
//   return nil
// }
//
// func (dst *RoleNameArray) Get() interface{} {
//   switch dst.Status {
//   case pgtype.Present:
//     return dst
//   case pgtype.Null:
//     return nil
//   default:
//     return dst.Status
//   }
// }
//
// func (src *RoleNameArray) AssignTo(dst interface{}) error {
//   switch src.Status {
//   case pgtype.Present:
//     switch v := dst.(type) {
//
//     case *[]string:
//       *v = make([]string, len(src.Elements))
//       for i := range src.Elements {
//         if err := src.Elements[i].AssignTo(&((*v)[i])); err != nil {
//           return err
//         }
//       }
//       return nil
//
//     default:
//       if nextDst, retry := pgtype.GetAssignToDstType(dst); retry {
//         return src.AssignTo(nextDst)
//       }
//     }
//   case pgtype.Null:
//     return pgtype.NullAssignTo(dst)
//   }
//
//   return fmt.Errorf("cannot decode %#v into %T", src, dst)
// }
//
// func (dst *RoleNameArray) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
//   if src == nil {
//     *dst = RoleNameArray{Status: pgtype.Null}
//     return nil
//   }
//
//   uta, err := pgtype.ParseUntypedTextArray(string(src))
//   if err != nil {
//     return err
//   }
//
//   var elements []RoleName
//
//   if len(uta.Elements) > 0 {
//     elements = make([]RoleName, len(uta.Elements))
//
//     for i, s := range uta.Elements {
//       var elem RoleName
//       var elemSrc []byte
//       if s != "NULL" {
//         elemSrc = []byte(s)
//       }
//       err = elem.DecodeText(ci, elemSrc)
//       if err != nil {
//         return err
//       }
//
//       elements[i] = elem
//     }
//   }
//
//   *dst = RoleNameArray{Elements: elements, Dimensions: uta.Dimensions, Status: pgtype.Present}
//
//   return nil
// }
//
// func (dst *RoleNameArray) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
//   if src == nil {
//     *dst = RoleNameArray{Status: pgtype.Null}
//     return nil
//   }
//
//   var arrayHeader pgtype.ArrayHeader
//   rp, err := arrayHeader.DecodeBinary(ci, src)
//   if err != nil {
//     return err
//   }
//
//   if len(arrayHeader.Dimensions) == 0 {
//     *dst = RoleNameArray{Dimensions: arrayHeader.Dimensions, Status: pgtype.Present}
//     return nil
//   }
//
//   elementCount := arrayHeader.Dimensions[0].Length
//   for _, d := range arrayHeader.Dimensions[1:] {
//     elementCount *= d.Length
//   }
//
//   elements := make([]RoleName, elementCount)
//
//   for i := range elements {
//     elemLen := int(int32(binary.BigEndian.Uint32(src[rp:])))
//     rp += 4
//     var elemSrc []byte
//     if elemLen >= 0 {
//       elemSrc = src[rp : rp+elemLen]
//       rp += elemLen
//     }
//     err = elements[i].DecodeBinary(ci, elemSrc)
//     if err != nil {
//       return err
//     }
//   }
//
//   *dst = RoleNameArray{Elements: elements, Dimensions: arrayHeader.Dimensions, Status: pgtype.Present}
//   return nil
// }
//
// func (src *RoleNameArray) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
//   switch src.Status {
//   case pgtype.Null:
//     return nil, nil
//   case pgtype.Undefined:
//     return nil, errUndefined
//   }
//
//   if len(src.Dimensions) == 0 {
//     return append(buf, '{', '}'), nil
//   }
//
//   buf = pgtype.EncodeTextArrayDimensions(buf, src.Dimensions)
//
//   // dimElemCounts is the multiples of elements that each array lies on. For
//   // example, a single dimension array of length 4 would have a dimElemCounts of
//   // [4]. A multi-dimensional array of lengths [3,5,2] would have a
//   // dimElemCounts of [30,10,2]. This is used to simplify when to render a '{'
//   // or '}'.
//   dimElemCounts := make([]int, len(src.Dimensions))
//   dimElemCounts[len(src.Dimensions)-1] = int(src.Dimensions[len(src.Dimensions)-1].Length)
//   for i := len(src.Dimensions) - 2; i > -1; i-- {
//     dimElemCounts[i] = int(src.Dimensions[i].Length) * dimElemCounts[i+1]
//   }
//
//   inElemBuf := make([]byte, 0, 32)
//   for i, elem := range src.Elements {
//     if i > 0 {
//       buf = append(buf, ',')
//     }
//
//     for _, dec := range dimElemCounts {
//       if i%dec == 0 {
//         buf = append(buf, '{')
//       }
//     }
//
//     elemBuf, err := elem.EncodeText(ci, inElemBuf)
//     if err != nil {
//       return nil, err
//     }
//     if elemBuf == nil {
//       buf = append(buf, `"NULL"`...)
//     } else {
//       buf = append(buf, pgtype.QuoteArrayElementIfNeeded(string(elemBuf))...)
//     }
//
//     for _, dec := range dimElemCounts {
//       if (i+1)%dec == 0 {
//         buf = append(buf, '}')
//       }
//     }
//   }
//
//   return buf, nil
// }
//
// func (src *RoleNameArray) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
//   switch src.Status {
//   case pgtype.Null:
//     return nil, nil
//   case pgtype.Undefined:
//     return nil, errUndefined
//   }
//
//   arrayHeader := pgtype.ArrayHeader{
//     Dimensions: src.Dimensions,
//   }
//
//   if dt, ok := ci.DataTypeForName("text"); ok {
//     arrayHeader.ElementOID = int32(dt.OID)
//   } else {
//     return nil, fmt.Errorf("unable to find oid for type name %v", "text")
//   }
//
//   for i := range src.Elements {
//     if src.Elements[i].Status == pgtype.Null {
//       arrayHeader.ContainsNull = true
//       break
//     }
//   }
//
//   buf = arrayHeader.EncodeBinary(ci, buf)
//
//   for i := range src.Elements {
//     sp := len(buf)
//     buf = pgio.AppendInt32(buf, -1)
//
//     elemBuf, err := src.Elements[i].EncodeBinary(ci, buf)
//     if err != nil {
//       return nil, err
//     }
//     if elemBuf != nil {
//       buf = elemBuf
//       pgio.SetInt32(buf[sp:], int32(len(buf[sp:])-4))
//     }
//   }
//
//   return buf, nil
// }
//
// // Scan implements the database/sql Scanner interface.
// func (dst *RoleNameArray) Scan(src interface{}) error {
//   if src == nil {
//     return dst.DecodeText(nil, nil)
//   }
//
//   switch src := src.(type) {
//   case string:
//     return dst.DecodeText(nil, []byte(src))
//   case []byte:
//     srcCopy := make([]byte, len(src))
//     copy(srcCopy, src)
//     return dst.DecodeText(nil, srcCopy)
//   }
//
//   return fmt.Errorf("cannot scan %T", src)
// }
//
// // Value implements the database/sql/driver Valuer interface.
// func (src *RoleNameArray) Value() (driver.Value, error) {
//   buf, err := src.EncodeText(nil, nil)
//   if err != nil {
//     return nil, err
//   }
//   if buf == nil {
//     return nil, nil
//   }
//
//   return string(buf), nil
// }

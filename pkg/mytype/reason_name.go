package mytype

// import (
//   "database/sql/driver"
//   "fmt"
//   "strings"
//
//   "github.com/jackc/pgx/pgtype"
// )
//
// type ReasonNameValue int
//
// const (
//   AuthorReason ReasonNameValue = iota
//   CommentReason
//   EnrolledReason
//   MentionReason
// )
//
// func (src ReasonNameValue) String() string {
//   switch src {
//   case AuthorReason:
//     return "author"
//   case CommentReason:
//     return "comment"
//   case EnrolledReason:
//     return "enrolled"
//   case MentionReason:
//     return "mention"
//   default:
//     return "unknown"
//   }
// }
//
// type ReasonName struct {
//   Status pgtype.Status
//   Name   ReasonNameValue
// }
//
// func NewReasonName(v ReasonNameValue) ReasonName {
//   return ReasonName{
//     Status: pgtype.Present,
//     Name:   v,
//   }
// }
//
// func ParseReasonName(s string) (ReasonName, error) {
//   switch strings.ToLower(s) {
//   case "author":
//     return ReasonName{
//       Status: pgtype.Present,
//       Name:   AuthorReason,
//     }, nil
//   case "comment":
//     return ReasonName{
//       Status: pgtype.Present,
//       Name:   CommentReason,
//     }, nil
//   case "enrolled":
//     return ReasonName{
//       Status: pgtype.Present,
//       Name:   EnrolledReason,
//     }, nil
//   case "mention":
//     return ReasonName{
//       Status: pgtype.Present,
//       Name:   MentionReason,
//     }, nil
//   default:
//     var o ReasonName
//     return o, fmt.Errorf("invalid ReasonName: %q", s)
//   }
// }
//
// func (src *ReasonName) String() string {
//   return src.Name.String()
// }
//
// func (dst *ReasonName) Set(src interface{}) error {
//   if src == nil {
//     *dst = ReasonName{Status: pgtype.Null}
//   }
//   switch value := src.(type) {
//   case ReasonName:
//     *dst = value
//     dst.Status = pgtype.Present
//   case *ReasonName:
//     *dst = *value
//     dst.Status = pgtype.Present
//   case ReasonNameValue:
//     dst.Name = value
//     dst.Status = pgtype.Present
//   case *ReasonNameValue:
//     dst.Name = *value
//     dst.Status = pgtype.Present
//   case string:
//     t, err := ParseReasonName(value)
//     if err != nil {
//       return err
//     }
//     *dst = t
//   case *string:
//     t, err := ParseReasonName(*value)
//     if err != nil {
//       return err
//     }
//     *dst = t
//   case []byte:
//     t, err := ParseReasonName(string(value))
//     if err != nil {
//       return err
//     }
//     *dst = t
//   default:
//     return fmt.Errorf("cannot convert %v to ReasonName", value)
//   }
//
//   return nil
// }
//
// func (src *ReasonName) Get() interface{} {
//   switch src.Status {
//   case pgtype.Present:
//     return src
//   case pgtype.Null:
//     return nil
//   default:
//     return src.Status
//   }
// }
//
// func (src *ReasonName) AssignTo(dst interface{}) error {
//   switch src.Status {
//   case pgtype.Present:
//     switch v := dst.(type) {
//     case *string:
//       *v = src.Name.String()
//       return nil
//     case *[]byte:
//       *v = make([]byte, len(src.Name.String()))
//       copy(*v, src.Name.String())
//       return nil
//     default:
//       if nextDst, retry := pgtype.GetAssignToDstType(dst); retry {
//         return src.AssignTo(nextDst)
//       }
//     }
//   case pgtype.Null:
//     return pgtype.NullAssignTo(dst)
//   }
//
//   return fmt.Errorf("cannot decode %v into %T", src, dst)
// }
//
// func (dst *ReasonName) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
//   if src == nil {
//     *dst = ReasonName{Status: pgtype.Null}
//     return nil
//   }
//
//   t, err := ParseReasonName(string(src))
//   if err != nil {
//     return err
//   }
//   *dst = t
//   return nil
// }
//
// func (dst *ReasonName) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
//   return dst.DecodeText(ci, src)
// }
//
// func (src *ReasonName) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
//   switch src.Status {
//   case pgtype.Null:
//     return nil, nil
//   case pgtype.Undefined:
//     return nil, errUndefined
//   }
//
//   return append(buf, src.Name.String()...), nil
// }
//
// func (src *ReasonName) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
//   return src.EncodeText(ci, buf)
// }
//
// // Scan implements the database/sql Scanner interface.
// func (dst *ReasonName) Scan(src interface{}) error {
//   if src == nil {
//     *dst = ReasonName{Status: pgtype.Null}
//     return nil
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
// func (src *ReasonName) Value() (driver.Value, error) {
//   switch src.Status {
//   case pgtype.Present:
//     return src.Name.String(), nil
//   case pgtype.Null:
//     return nil, nil
//   default:
//     return nil, errUndefined
//   }
// }

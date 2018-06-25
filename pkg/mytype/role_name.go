package mytype

// import (
//   "database/sql/driver"
//   "fmt"
//   "strings"
//
//   "github.com/jackc/pgx/pgtype"
// )
//
// type RoleNameValue int
//
// const (
//   AdminRole RoleNameValue = iota
//   MemberRole
//   OwnerRole
//   UserRole
// )
//
// func (r RoleNameValue) String() string {
//   switch r {
//   case AdminRole:
//     return "ADMIN"
//   case MemberRole:
//     return "MEMBER"
//   case OwnerRole:
//     return "OWNER"
//   case UserRole:
//     return "USER"
//   default:
//     return "unknown"
//   }
// }
//
// type RoleName struct {
//   Status pgtype.Status
//   Name   RoleNameValue
// }
//
// func NewRoleName(v RoleNameValue) RoleName {
//   return RoleName{
//     Status: pgtype.Present,
//     Name:   v,
//   }
// }
//
// func ParseRoleName(s string) (RoleName, error) {
//   switch strings.ToUpper(s) {
//   case "ADMIN":
//     return RoleName{
//       Status: pgtype.Present,
//       Name:   AdminRole,
//     }, nil
//   case "MEMBER":
//     return RoleName{
//       Status: pgtype.Present,
//       Name:   MemberRole,
//     }, nil
//   case "OWNER":
//     return RoleName{
//       Status: pgtype.Present,
//       Name:   OwnerRole,
//     }, nil
//   case "USER":
//     return RoleName{
//       Status: pgtype.Present,
//       Name:   UserRole,
//     }, nil
//   default:
//     var o RoleName
//     return o, fmt.Errorf("invalid RoleName: %q", s)
//   }
// }
//
// func (src *RoleName) String() string {
//   return src.Name.String()
// }
//
// func (dst *RoleName) Set(src interface{}) error {
//   if src == nil {
//     *dst = RoleName{Status: pgtype.Null}
//   }
//   switch value := src.(type) {
//   case RoleName:
//     *dst = value
//     dst.Status = pgtype.Present
//   case *RoleName:
//     *dst = *value
//     dst.Status = pgtype.Present
//   case RoleNameValue:
//     dst.Name = value
//     dst.Status = pgtype.Present
//   case *RoleNameValue:
//     dst.Name = *value
//     dst.Status = pgtype.Present
//   case string:
//     t, err := ParseRoleName(value)
//     if err != nil {
//       return err
//     }
//     *dst = t
//   case *string:
//     t, err := ParseRoleName(*value)
//     if err != nil {
//       return err
//     }
//     *dst = t
//   case []byte:
//     t, err := ParseRoleName(string(value))
//     if err != nil {
//       return err
//     }
//     *dst = t
//   default:
//     return fmt.Errorf("cannot convert %v to RoleName", value)
//   }
//
//   return nil
// }
//
// func (src *RoleName) Get() interface{} {
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
// func (src *RoleName) AssignTo(dst interface{}) error {
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
// func (dst *RoleName) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
//   if src == nil {
//     *dst = RoleName{Status: pgtype.Null}
//     return nil
//   }
//
//   t, err := ParseRoleName(string(src))
//   if err != nil {
//     return err
//   }
//   *dst = t
//   return nil
// }
//
// func (dst *RoleName) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
//   return dst.DecodeText(ci, src)
// }
//
// func (src *RoleName) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
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
// func (src *RoleName) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
//   return src.EncodeText(ci, buf)
// }
//
// // Scan implements the database/sql Scanner interface.
// func (dst *RoleName) Scan(src interface{}) error {
//   if src == nil {
//     *dst = RoleName{Status: pgtype.Null}
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
// func (src *RoleName) Value() (driver.Value, error) {
//   switch src.Status {
//   case pgtype.Present:
//     return src.Name.String(), nil
//   case pgtype.Null:
//     return nil, nil
//   default:
//     return nil, errUndefined
//   }
// }

package repo

// import (
//   "net/http"
//   "time"
//
//   "github.com/fatih/structs"
//   "github.com/iancoleman/strcase"
//   "github.com/marksauter/markus-ninja-api/pkg/data"
//   "github.com/marksauter/markus-ninja-api/pkg/loader"
//   "github.com/marksauter/markus-ninja-api/pkg/mylog"
//   "github.com/marksauter/markus-ninja-api/pkg/perm"
// )
//
// type UserEmailPermit struct {
//   checkFieldPermission FieldPermissionFunc
//   userEmail            *data.UserEmail
// }
//
// func (r *UserEmailPermit) PreCheckPermissions() error {
//   for _, f := range structs.Fields(r.userEmail) {
//     if !f.IsZero() {
//       if ok := r.checkFieldPermission(strcase.ToSnake(f.Name())); !ok {
//         return ErrAccessDenied
//       }
//     }
//   }
//   return nil
// }
//
// func (r *UserEmailPermit) CreatedAt() (time.Time, error) {
//   if ok := r.checkFieldPermission("created_at"); !ok {
//     return time.Time{}, ErrAccessDenied
//   }
//   return r.userEmail.CreatedAt.Time, nil
// }
//
// func (r *UserEmailPermit) EmailId() (string, error) {
//   if ok := r.checkFieldPermission("email_id"); !ok {
//     return "", ErrAccessDenied
//   }
//   return r.userEmail.EmailId.String, nil
// }
//
// func (r *UserEmailPermit) Type() (string, error) {
//   if ok := r.checkFieldPermission("type"); !ok {
//     return "", ErrAccessDenied
//   }
//   return r.userEmail.Type.String(), nil
// }
//
// func (r *UserEmailPermit) UpdatedAt() (time.Time, error) {
//   if ok := r.checkFieldPermission("updated_at"); !ok {
//     return time.Time{}, ErrAccessDenied
//   }
//   return r.userEmail.UpdatedAt.Time, nil
// }
//
// func (r *UserEmailPermit) UserId() (string, error) {
//   if ok := r.checkFieldPermission("user_id"); !ok {
//     return "", ErrAccessDenied
//   }
//   return r.userEmail.UserId.String, nil
// }
//
// func (r *UserEmailPermit) VerifiedAt() (time.Time, error) {
//   if ok := r.checkFieldPermission("verified_at"); !ok {
//     return time.Time{}, ErrAccessDenied
//   }
//   return r.userEmail.VerifiedAt.Time, nil
// }
//
// func NewUserEmailRepo(
//   perms *PermRepo,
//   svc *data.UserEmailService,
// ) *UserEmailRepo {
//   return &UserEmailRepo{
//     perms: perms,
//     svc:   svc,
//   }
// }
//
// type UserEmailRepo struct {
//   load  *loader.UserEmailLoader
//   perms *PermRepo
//   svc   *data.UserEmailService
// }
//
// func (r *UserEmailRepo) Open(ctx context.Context) error {
// err := r.perms.Open(ctx)
// if err != nil {
//   return err
// }
// if r.load == nil {
//   r.load = loader.NewUserEmailLoader(r.svc)
// }
// return nil
// }
//
// func (r *UserEmailRepo) Close() {
//   r.load = nil
// }
//
// // Service methods
//
// func (r *UserEmailRepo) Create(userEmail *data.UserEmail) (*UserEmailPermit, error) {
//   fieldPermFn, err := r.perms.Check(perm.CreateUserEmail)
//   if err != nil {
//     return nil, err
//   }
//   if r.load == nil {
//     mylog.Log.Error("userEmail connection closed")
//     return nil, ErrConnClosed
//   }
//   userEmailPermit := &UserEmailPermit{fieldPermFn, userEmail}
//   err := userEmailPermit.PreCheckPermissions()
//   if err != nil {
//     return nil, err
//   }
//   err = r.svc.Create(userEmail)
//   if err != nil {
//     return nil, err
//   }
//   return userEmailPermit, nil
// }
//
// func (r *UserEmailRepo) Get(userId, emailId string) (*UserEmailPermit, error) {
//   fieldPermFn, err := r.perms.Check(perm.ReadUserEmail)
//   if err != nil {
//     return nil, err
//   }
//   if r.load == nil {
//     mylog.Log.Error("userEmail connection closed")
//     return nil, ErrConnClosed
//   }
//   userEmail, err := r.load.Get(userId, emailId)
//   if err != nil {
//     return nil, err
//   }
//   return &UserEmailPermit{fieldPermFn, userEmail}, nil
// }
//
// // Middleware
// func (r *UserEmailRepo) Use(h http.Handler) http.Handler {
//   return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// r.Open(req.Context())
//     defer r.Close()
//     h.ServeHTTP(rw, req)
//   })
// }

package repo

// import (
//   "context"
//   "net/http"
//   "time"
//
//   "github.com/fatih/structs"
//   "github.com/marksauter/markus-ninja-api/pkg/data"
//   "github.com/marksauter/markus-ninja-api/pkg/loader"
//   "github.com/marksauter/markus-ninja-api/pkg/mylog"
//   "github.com/marksauter/markus-ninja-api/pkg/oid"
//   "github.com/marksauter/markus-ninja-api/pkg/perm"
// )
//
// type EmailPermit struct {
//   checkFieldPermission FieldPermissionFunc
//   email                *data.Email
// }
//
// func (r *EmailPermit) Get() *data.Email {
//   email := r.email
//   fields := structs.Fields(email)
//   for _, f := range fields {
//     name := f.Tag("db")
//     if ok := r.checkFieldPermission(name); !ok {
//       f.Zero()
//     }
//   }
//   return email
// }
//
// func (r *EmailPermit) CreatedAt() (time.Time, error) {
//   if ok := r.checkFieldPermission("created_at"); !ok {
//     return time.Time{}, ErrAccessDenied
//   }
//   return r.email.CreatedAt.Time, nil
// }
//
// func (r *EmailPermit) ID() (*oid.OID, error) {
//   if ok := r.checkFieldPermission("id"); !ok {
//     return nil, ErrAccessDenied
//   }
//   return &r.email.Id, nil
// }
//
// func (r *EmailPermit) Value() (string, error) {
//   if ok := r.checkFieldPermission("value"); !ok {
//     return "", ErrAccessDenied
//   }
//   return r.email.Value.String, nil
// }
//
// func NewEmailRepo(perms *PermRepo, svc *data.EmailService) *EmailRepo {
//   return &EmailRepo{
//     perms: perms,
//     svc:   svc,
//   }
// }
//
// type EmailRepo struct {
//   load  *loader.EmailLoader
//   perms *PermRepo
//   svc   *data.EmailService
// }
//
// func (r *EmailRepo) Open(ctx context.Context) error {
//   err := r.perms.Open(ctx)
//   if err != nil {
//     return err
//   }
//   if r.load == nil {
//     r.load = loader.NewEmailLoader(r.svc)
//   }
//   return nil
// }
//
// func (r *EmailRepo) Close() {
//   r.load = nil
// }
//
// func (r *EmailRepo) CheckConnection() error {
//   if r.load == nil {
//     mylog.Log.Error("email connection closed")
//     return ErrConnClosed
//   }
//   return nil
// }
//
// // Service methods
//
// func (r *EmailRepo) Create(email *data.Email) (*EmailPermit, error) {
//   if err := r.CheckConnection(); err != nil {
//     return nil, err
//   }
//   if _, err := r.perms.Check2(perm.Create, email); err != nil {
//     return nil, err
//   }
//   if err := r.svc.Create(email); err != nil {
//     return nil, err
//   }
//   fieldPermFn, err := r.perms.Check2(perm.Read, email)
//   if err != nil {
//     return nil, err
//   }
//   return &EmailPermit{fieldPermFn, email}, nil
// }
//
// func (r *EmailRepo) Get(id string) (*EmailPermit, error) {
//   if err := r.CheckConnection(); err != nil {
//     return nil, err
//   }
//   email, err := r.load.Get(id)
//   if err != nil {
//     return nil, err
//   }
//   fieldPermFn, err := r.perms.Check2(perm.Read, email)
//   if err != nil {
//     return nil, err
//   }
//   return &EmailPermit{fieldPermFn, email}, nil
// }
//
// // Middleware
// func (r *EmailRepo) Use(h http.Handler) http.Handler {
//   return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
//     r.Open(req.Context())
//     defer r.Close()
//     h.ServeHTTP(rw, req)
//   })
// }

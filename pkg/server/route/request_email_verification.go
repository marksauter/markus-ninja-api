package route

// import (
//   "net/http"
//
//   "github.com/gorilla/mux"
//   "github.com/marksauter/markus-ninja-api/pkg/data"
//   "github.com/marksauter/markus-ninja-api/pkg/myhttp"
//   "github.com/marksauter/markus-ninja-api/pkg/mylog"
//   "github.com/marksauter/markus-ninja-api/pkg/server/middleware"
//   "github.com/marksauter/markus-ninja-api/pkg/service"
//   "github.com/rs/cors"
// )
//
// func RequestEmailVerification(svcs *service.Services) http.Handler {
//   requestAccountVerificationHandler := RequestEmailVerificationHandler{
//     Svcs: svcs,
//   }
//   return middleware.CommonMiddleware.Append(
//     requestEmailVerificationCors.Handler,
//   ).Then(requestAccountVerificationHandler)
// }
//
// var requestEmailVerificationCors = cors.New(cors.Options{
//   AllowedHeaders: []string{"Content-Type"},
//   AllowedMethods: []string{http.MethodOptions, http.MethodPost},
//   AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
// })
//
// type RequestEmailVerificationHandler struct {
//   Svcs *service.Services
// }
//
// func (h RequestEmailVerificationHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
//   if req.Method != http.MethodPost {
//     response := myhttp.MethodNotAllowedResponse(req.Method)
//     myhttp.WriteResponseTo(rw, response)
//     return
//   }
//
//   routeVars := mux.Vars(req)
//
//   // var verify struct {
//   //   Email string `json:"email"`
//   // }
//   //
//   // err := myhttp.UnmarshalRequestBody(req, &verify)
//   // if err != nil {
//   //   response := myhttp.InvalidRequestErrorResponse(err.Error())
//   //   myhttp.WriteResponseTo(rw, response)
//   //   return
//   // }
//   // if verify.Email == "" {
//   //   response := &myhttp.ErrorResponse{
//   //     Error:            myhttp.InvalidCredentials,
//   //     ErrorDescription: "The request email was invalid",
//   //   }
//   //   myhttp.WriteResponseTo(rw, response)
//   // }
//   // if len(verify.Email) > 40 {
//   //   response := &myhttp.ErrorResponse{
//   //     Error:            myhttp.InvalidCredentials,
//   //     ErrorDescription: "The request email must be less than or equal to 40 characters",
//   //   }
//   //   myhttp.WriteResponseTo(rw, response)
//   //   return
//   // }
//
//   login := routeVars["login"]
//   user, err := h.Svcs.User.GetCredentialsByLogin(login)
//   if err != nil {
//     response := myhttp.InternalServerErrorResponse(err.Error())
//     myhttp.WriteResponseTo(rw, response)
//     return
//   }
//
//   emailId := routeVars["id"]
//   email, err := h.Svcs.Email.GetByPK(emailId)
//   if err != nil {
//     response := myhttp.InternalServerErrorResponse(err.Error())
//     myhttp.WriteResponseTo(rw, response)
//     return
//   }
//   avt := &data.EmailVerificationTokenModel{}
//   avt.EmailId.Set(email.Id)
//   avt.UserId.Set(user.Id)
//
//   err = h.Svcs.AVT.Create(avt)
//   if err != nil {
//     response := myhttp.InternalServerErrorResponse(err.Error())
//     myhttp.WriteResponseTo(rw, response)
//     return
//   }
//
//   if user == nil {
//     mylog.Log.WithField(
//       "email",
//       verify.Email,
//     ).Warn("Account verification requested for missing email")
//     return
//   }
//
//   err = h.Svcs.Mail.SendEmailVerificationMail(verify.Email, login, avt.Token.String)
//   if err != nil {
//     response := myhttp.InternalServerErrorResponse(err.Error())
//     myhttp.WriteResponseTo(rw, response)
//     return
//   }
//
//   rw.WriteHeader(http.StatusOK)
//   return
// }

package route

//
// import (
//   "net/http"
//   "time"
//
//   "github.com/justinas/alice"
//   "github.com/marksauter/markus-ninja-api/pkg/model"
//   "github.com/marksauter/markus-ninja-api/pkg/myhttp"
//   "github.com/marksauter/markus-ninja-api/pkg/myjwt"
//   "github.com/marksauter/markus-ninja-api/pkg/repo"
//   "github.com/marksauter/markus-ninja-api/pkg/service"
//   "github.com/rs/cors"
// )
//
// type SignupSuccessResponse struct {
//   AccessToken string              `json:"access_token,omitempty"`
//   ExpiresIn   myjwt.UnixTimestamp `json:"expires_in,omitempty"`
// }
//
// func (r *SignupSuccessResponse) StatusHTTP() int {
//   return http.StatusOK
// }
//
// var Signup = alice.New(SignupCors.Handler).Then(SignupHandler{})
//
// var SignupCors = cors.New(cors.Options{
//   AllowedHeaders: []string{"Content-Type"},
//   AllowedMethods: []string{http.MethodOptions, http.MethodPost},
//   AllowedOrigins: []string{"*"},
// })
//
// type SignupHandler struct {
//   AuthSvc  *service.AuthService
//   UserRepo *repo.UserRepo
// }
//
// func (h SignupHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
//   rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
//   rw.Header().Set("Cache-Control", "no-store")
//   rw.Header().Set("Pragma", "no-cache")
//
//   if req.Method != http.MethodPost {
//     response := myhttp.MethodNotAllowedResponse(req.Method)
//     myhttp.WriteResponseTo(rw, response)
//     return
//   }
//   err := req.ParseForm()
//   if err != nil {
//     response := myhttp.InternalServerErrorResponse(err.Error())
//     myhttp.WriteResponseTo(rw, response)
//     return
//   }
//
//   user := new(model.User)
//
//   exp := time.Now().Add(time.Hour * time.Duration(24)).Unix()
//   payload := myjwt.Payload{Exp: exp, Iat: time.Now().Unix(), Sub: user.Id}
//   jwt, err := h.AuthSvc.SignJWT(&payload)
//   if err != nil {
//     response := myhttp.InternalServerErrorResponse(err.Error())
//     myhttp.WriteResponseTo(rw, response)
//     return
//   }
//
//   response := SignupSuccessResponse{
//     AccessToken: jwt.String(),
//     ExpiresIn:   jwt.Payload.Exp,
//   }
//   myhttp.WriteResponseTo(rw, &response)
// }

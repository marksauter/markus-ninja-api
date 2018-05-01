package route

import (
	"net/http"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/rs/cors"
)

func Signup(authSvc *data.AuthService, repos *repo.Repos) http.Handler {
	authMiddleware := middleware.Authenticate{
		AuthSvc: authSvc,
		Repos:   repos,
	}
	signupHandler := SignupHandler{
		AuthSvc: authSvc,
		Repos:   repos,
	}
	return middleware.CommonMiddleware.Append(
		SignupCors.Handler,
		repos.Use,
		authMiddleware.Use,
	).Then(signupHandler)
}

var SignupCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"*"},
})

type SignupHandler struct {
	AuthSvc *data.AuthService
	Repos   *repo.Repos
}

func (h SignupHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}
	var registration struct {
		Email    string `json:"email"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := myhttp.UnmarshalRequestBody(req, &registration)
	if err != nil {
		response := myhttp.BadRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if registration.Email == "" {
		response := myhttp.BadRequestErrorResponse("invalid email")
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if registration.Login == "" {
		response := myhttp.BadRequestErrorResponse("invalid login")
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if registration.Password == "" {
		response := myhttp.BadRequestErrorResponse("invalid password")
		myhttp.WriteResponseTo(rw, response)
		return
	}

	password := passwd.New(registration.Password)
	if err := password.CheckStrength(passwd.VeryWeak); err != nil {
		response := myhttp.PasswordStrengthErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	user := &data.UserModel{}
	user.Login.Set(registration.Login)
	user.Password.Set(password.Hash())
	user.PrimaryEmail.Set(registration.Email)

	createUserPerm, err := h.Repos.Perm().GetQueryPermission(perm.CreateUser)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	h.Repos.User().AddPermission(*createUserPerm)
	_, err = h.Repos.User().Create(user)
	if err != nil {
		var response *myhttp.ErrorResponse
		if dfErr, ok := err.(data.DataFieldError); ok {
			switch dfErr.Code {
			case data.DuplicateField:
				if dfErr.Field == "login" {
					response = myhttp.UsernameExistsResponse()
				} else if dfErr.Field == "primary_email" {
					response = myhttp.UserExistsResponse()
				} else {
					response = myhttp.InternalServerErrorResponse(dfErr.Error())
				}
			case data.RequiredField:
				response = myhttp.BadRequestErrorResponse(dfErr.Error())
			default:
				response = myhttp.InternalServerErrorResponse(dfErr.Error())
			}
		} else {
			response = myhttp.InternalServerErrorResponse(err.Error())
		}
		myhttp.WriteResponseTo(rw, response)
		return
	}

	exp := time.Now().Add(time.Hour * time.Duration(24)).Unix()
	payload := myjwt.Payload{Exp: exp, Iat: time.Now().Unix(), Sub: user.Id.String}
	jwt, err := h.AuthSvc.SignJWT(&payload)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	response := SignupSuccessResponse{
		AccessToken: jwt.String(),
		ExpiresIn:   jwt.Payload.Exp,
	}
	myhttp.WriteResponseTo(rw, &response)
}

type SignupSuccessResponse struct {
	AccessToken string              `json:"access_token,omitempty"`
	ExpiresIn   myjwt.UnixTimestamp `json:"expires_in,omitempty"`
}

func (r *SignupSuccessResponse) StatusHTTP() int {
	return http.StatusOK
}

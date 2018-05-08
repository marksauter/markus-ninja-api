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
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

func Signup(svcs *service.Services, repos *repo.Repos) http.Handler {
	authMiddleware := middleware.Authenticate{
		Svcs:  svcs,
		Repos: repos,
	}
	signupHandler := SignupHandler{
		Svcs:  svcs,
		Repos: repos,
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
	Svcs  *service.Services
	Repos *repo.Repos
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
		Username string `json:"login"`
		Password string `json:"password"`
	}
	err := myhttp.UnmarshalRequestBody(req, &registration)
	if err != nil {
		response := myhttp.InvalidRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if registration.Email == "" {
		response := &myhttp.ErrorResponse{
			Error:            myhttp.InvalidCredentials,
			ErrorDescription: "The request email was invalid",
		}
		myhttp.WriteResponseTo(rw, response)
		return
	}
	if len(registration.Email) > 40 {
		response := &myhttp.ErrorResponse{
			Error:            myhttp.InvalidCredentials,
			ErrorDescription: "The request email must be less than or equal to 40 characters",
		}
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if registration.Username == "" {
		response := &myhttp.ErrorResponse{
			Error:            myhttp.InvalidCredentials,
			ErrorDescription: "The request username was invalid",
		}
		myhttp.WriteResponseTo(rw, response)
		return
	}
	if len(registration.Username) > 40 {
		response := &myhttp.ErrorResponse{
			Error:            myhttp.InvalidCredentials,
			ErrorDescription: "The request username must be less than or equal to 40 characters",
		}
		myhttp.WriteResponseTo(rw, response)
		return
	}

	password, err := passwd.New(registration.Password)
	if err != nil {
		response := myhttp.InvalidPasswordResponse()
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if err := password.CheckStrength(passwd.VeryWeak); err != nil {
		response := myhttp.PasswordStrengthErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	user := &data.User{}
	user.Login.Set(registration.Username)
	user.Password.Set(password.Hash())
	user.PrimaryEmail.Set(registration.Email)

	createUserPerm, err := h.Repos.Perm().GetQueryPermission(perm.CreateUser)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	h.Repos.User().AddPermission(createUserPerm)
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
				response = myhttp.InvalidRequestErrorResponse(dfErr.Error())
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
	jwt, err := h.Svcs.Auth.SignJWT(&payload)
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

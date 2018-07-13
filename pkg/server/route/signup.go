package route

import (
	"net/http"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

var SignupCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "http://localhost:*"},
})

type SignupHandler struct {
	AuthSvc *service.AuthService
	Db      data.Queryer
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

	u := &data.User{}
	u.Login.Set(registration.Username)
	u.PrimaryEmail.Set(registration.Email)

	if err := u.Password.Set(registration.Password); err != nil {
		response := myhttp.InvalidPasswordResponse()
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if err := u.Password.CheckStrength(mytype.VeryWeak); err != nil {
		response := myhttp.PasswordStrengthErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	user, err := data.CreateUser(h.Db, u)
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

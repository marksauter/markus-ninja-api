package route

import (
	"net/http"
	"time"

	"github.com/badoux/checkmail"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

func Token(svcs *service.Services, repos *repo.Repos) http.Handler {
	tokenHandler := TokenHandler{
		Svcs:  svcs,
		Repos: repos,
	}
	return middleware.CommonMiddleware.Append(
		tokenCors.Handler,
		repos.User().Use,
	).Then(tokenHandler)
}

var tokenCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
})

type TokenHandler struct {
	Svcs  *service.Services
	Repos *repo.Repos
}

func (h TokenHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}
	validationOutput, err := myhttp.ValidateBasicAuthHeader(req)
	if err == myhttp.ErrNoAuthHeader {
		validationOutput = &myhttp.ValidateBasicAuthHeaderOutput{
			Login:    "guest",
			Password: "guest",
		}
	} else if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	verificationInput := repo.VerifyCredentialsInput{
		Password: validationOutput.Password,
	}
	err = checkmail.ValidateFormat(validationOutput.Login)
	if err != nil {
		verificationInput.Login = validationOutput.Login
	} else {
		verificationInput.Email = validationOutput.Login
	}

	user, err := h.Repos.User().VerifyCredentials(&verificationInput)
	if err != nil {
		response := myhttp.InvalidCredentialsErrorResponse()
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

	response := TokenSuccessResponse{
		AccessToken: jwt.String(),
		ExpiresIn:   jwt.Payload.Exp,
	}
	myhttp.WriteResponseTo(rw, &response)
}

type TokenSuccessResponse struct {
	AccessToken string              `json:"access_token,omitempty"`
	ExpiresIn   myjwt.UnixTimestamp `json:"expires_in,omitempty"`
}

func (r *TokenSuccessResponse) StatusHTTP() int {
	return http.StatusOK
}

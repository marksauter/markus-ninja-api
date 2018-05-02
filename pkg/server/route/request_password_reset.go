package route

import (
	"net/http"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/rs/cors"
)

func RequestPasswordReset(authSvc *data.AuthService, userRepo *repo.UserRepo) http.Handler {
	tokenHandler := RequestPasswordResetHandler{
		AuthSvc:  authSvc,
		UserRepo: userRepo,
	}
	return middleware.CommonMiddleware.Append(
		tokenCors.Handler,
		userRepo.Use,
	).Then(tokenHandler)
}

var requestPasswordResetCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
})

type RequestPasswordResetHandler struct {
	AuthSvc  *data.AuthService
	UserRepo *repo.UserRepo
}

func (h RequestPasswordResetHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}
	validationOutput, err := myhttp.ValidateBasicAuthHeader(req)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	verificationInput := repo.VerifyCredentialsInput{
		Login:    validationOutput.Login,
		Password: validationOutput.Password,
	}

	user, err := h.UserRepo.VerifyCredentials(&verificationInput)
	if err != nil {
		response := myhttp.InvalidCredentialsErrorResponse()
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

	response := RequestPasswordResetSuccessResponse{
		AccessRequestPasswordReset: jwt.String(),
		ExpiresIn:                  jwt.Payload.Exp,
	}
	myhttp.WriteResponseTo(rw, &response)
}

type RequestPasswordResetSuccessResponse struct {
	AccessRequestPasswordReset string              `json:"access_token,omitempty"`
	ExpiresIn                  myjwt.UnixTimestamp `json:"expires_in,omitempty"`
}

func (r *RequestPasswordResetSuccessResponse) StatusHTTP() int {
	return http.StatusOK
}

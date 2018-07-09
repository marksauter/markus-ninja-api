package route

import (
	"net/http"
	"time"

	"github.com/badoux/checkmail"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

func Token(db data.Queryer, svcs *service.Services) http.Handler {
	tokenHandler := TokenHandler{db, svcs}
	return middleware.CommonMiddleware.Append(
		tokenCors.Handler,
	).Then(tokenHandler)
}

var tokenCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
})

type TokenHandler struct {
	db   data.Queryer
	svcs *service.Services
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
	creds, err := myhttp.ValidateBasicAuthHeader(req)
	if err == myhttp.ErrNoAuthHeader {
		creds = &myhttp.ValidateBasicAuthHeaderOutput{
			Login:    "guest",
			Password: "guest",
		}
	} else if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	var user *data.User
	err = checkmail.ValidateFormat(creds.Login)
	if err != nil {
		user, err = data.GetUserCredentialsByLogin(h.db, creds.Login)
		if err != nil {
			response := myhttp.InvalidCredentialsErrorResponse()
			myhttp.WriteResponseTo(rw, response)
			return
		}
	} else {
		user, err = data.GetUserCredentialsByEmail(h.db, creds.Login)
		if err != nil {
			response := myhttp.InvalidCredentialsErrorResponse()
			myhttp.WriteResponseTo(rw, response)
			return
		}
	}

	if err = user.Password.CompareToPassword(creds.Password); err != nil {
		mylog.Log.WithError(err).Error("passwords do not match")
		response := myhttp.InvalidCredentialsErrorResponse()
		myhttp.WriteResponseTo(rw, response)
		return
	}

	exp := time.Now().Add(time.Hour * time.Duration(24)).Unix()
	payload := myjwt.Payload{Exp: exp, Iat: time.Now().Unix(), Sub: user.Id.String}
	jwt, err := h.svcs.Auth.SignJWT(&payload)
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
	return
}

type TokenSuccessResponse struct {
	AccessToken string              `json:"access_token,omitempty"`
	ExpiresIn   myjwt.UnixTimestamp `json:"expires_in,omitempty"`
}

func (r *TokenSuccessResponse) StatusHTTP() int {
	return http.StatusOK
}

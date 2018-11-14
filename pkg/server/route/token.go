package route

import (
	"errors"
	"net/http"
	"time"

	"github.com/badoux/checkmail"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/cors"
)

type TokenHandler struct {
	AuthSvc *service.AuthService
	Conf    *myconf.Config
	Db      data.Queryer
}

func (h TokenHandler) Cors() *cors.Cors {
	branch := util.GetRequiredEnv("BRANCH")
	allowedOrigins := []string{"ma.rkus.ninja"}
	if branch != "production" {
		allowedOrigins = append(allowedOrigins, "http://localhost:*")
	}

	return cors.New(cors.Options{
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		AllowedMethods:   []string{http.MethodOptions, http.MethodGet},
		AllowedOrigins:   allowedOrigins,
		// Debug: true,
	})
}

func (h TokenHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.AuthSvc == nil || h.Conf == nil || h.Db == nil {
		err := errors.New("route inproperly setup")
		mylog.Log.WithError(err).Error(util.Trace(""))
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	if req.Method != http.MethodGet {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}
	creds, err := myhttp.ValidateBasicAuthHeader(req)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	var user *data.User
	err = checkmail.ValidateFormat(creds.Login)
	if err != nil {
		user, err = data.GetUserCredentialsByLogin(h.Db, creds.Login)
		if err != nil {
			response := myhttp.InvalidCredentialsErrorResponse()
			myhttp.WriteResponseTo(rw, response)
			return
		}
	} else {
		user, err = data.GetUserCredentialsByEmail(h.Db, creds.Login)
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
	payload := myjwt.Payload{Exp: exp, Iat: time.Now().Unix(), Sub: user.ID.String}
	jwt, err := h.AuthSvc.SignJWT(&payload)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	http.SetCookie(rw, &http.Cookie{
		Name:     "access_token",
		Value:    jwt.String(),
		Expires:  time.Unix(jwt.Payload.Exp, 0),
		HttpOnly: true,
		// Secure:   true,
	})
	return
}

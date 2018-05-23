package route

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

func ConfirmVerification(svcs *service.Services) http.Handler {
	verifyAcccountHandler := ConfirmVerificationHandler{
		Svcs: svcs,
	}
	return middleware.CommonMiddleware.Append(
		confirmVerificationCors.Handler,
	).Then(verifyAcccountHandler)
}

var confirmVerificationCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
})

type ConfirmVerificationHandler struct {
	Svcs *service.Services
}

func (h ConfirmVerificationHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	routeVars := mux.Vars(req)

	login := routeVars["login"]
	user, err := h.Svcs.User.GetByLogin(login)
	if err == data.ErrNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	emailId, err := oid.NewFromShort("User", routeVars["id"])
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	token := routeVars["token"]
	evt, err := h.Svcs.EVT.GetByPK(emailId.String, token)
	if err == data.ErrNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if evt.UserId.String != user.Id.String {
		mylog.Log.WithField(
			"login", user.Login.String,
		).Warn("user attempting to use another user's email verification token")
		rw.WriteHeader(http.StatusNotFound)
	}

	if evt.VerifiedAt.Status == pgtype.Present {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if evt.ExpiresAt.Time.Before(time.Now()) {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	err = h.Svcs.Role.GrantUser(evt.UserId.String, data.UserRole)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	err = evt.VerifiedAt.Set(time.Now())
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	err = h.Svcs.EVT.Update(evt)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.WriteHeader(http.StatusOK)
	return
}

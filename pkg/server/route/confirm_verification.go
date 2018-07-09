package route

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/rs/cors"
)

func ConfirmVerification(db data.Queryer) http.Handler {
	confirmVerificationHandler := ConfirmVerificationHandler{db}
	return middleware.CommonMiddleware.Append(
		confirmVerificationCors.Handler,
	).Then(confirmVerificationHandler)
}

var confirmVerificationCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
})

type ConfirmVerificationHandler struct {
	db data.Queryer
}

func (h ConfirmVerificationHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	routeVars := mux.Vars(req)

	login := routeVars["login"]
	user, err := data.GetUserByLogin(h.db, login)
	if err == data.ErrNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	emailId, err := mytype.NewOIDFromShort("Email", routeVars["id"])
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	token := routeVars["token"]
	evt, err := data.GetEVT(h.db, emailId.String, token)
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
		mylog.Log.Warn("token has already by used")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if evt.ExpiresAt.Time.Before(time.Now()) {
		mylog.Log.Warn("token has expired")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	err = data.GrantUserRole(h.db, evt.UserId.String, data.UserRole)
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
	evt, err = data.UpdateEVT(h.db, evt)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	email, err := data.GetEmail(h.db, evt.EmailId.String)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if err := email.VerifiedAt.Set(time.Now()); err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if _, err := data.UpdateEmail(h.db, email); err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.WriteHeader(http.StatusOK)
	return
}

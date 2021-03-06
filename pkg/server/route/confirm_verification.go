package route

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/cors"
)

type ConfirmVerificationHandler struct {
	Conf *myconf.Config
	Db   data.Queryer
}

func (h ConfirmVerificationHandler) Cors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedHeaders: []string{"Content-Type"},
		AllowedMethods: []string{http.MethodOptions, http.MethodPost},
		AllowedOrigins: []string{h.Conf.ClientURL},
	})
}

func (h ConfirmVerificationHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.Conf == nil || h.Db == nil {
		err := errors.New("route inproperly setup")
		mylog.Log.WithError(err).Error(util.Trace(""))
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if req.Method != http.MethodGet {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	routeVars := mux.Vars(req)

	login := routeVars["login"]
	user, err := data.GetUserByLogin(h.Db, login)
	if err == data.ErrNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	emailID, err := mytype.NewOIDFromShort("Email", routeVars["id"])
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	token := routeVars["token"]
	evt, err := data.GetEVT(h.Db, emailID.String, token)
	if err == data.ErrNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if evt.UserID.String != user.ID.String {
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

	err = data.GrantUserRoles(h.Db, evt.UserID.String, data.UserRole)
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
	evt, err = data.UpdateEVT(h.Db, evt)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	email, err := data.GetEmail(h.Db, evt.EmailID.String)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if err := email.VerifiedAt.Set(time.Now()); err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if _, err := data.UpdateEmail(h.Db, email); err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	http.Redirect(rw, req, h.Conf.ClientURL, http.StatusSeeOther)
	return
}

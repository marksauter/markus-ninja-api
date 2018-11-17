package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/xid"
)

type Middleware interface {
	Use(context.Context, http.Handler) http.Handler
}

var CommonMiddleware = alice.New(
	mylog.Log.AccessMiddleware,
	handlers.RecoveryHandler(),
)

type AddContext struct {
	Ctx context.Context
}

func (a *AddContext) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(rw, req.WithContext(a.Ctx))
	})
}

type Authenticate struct {
	AuthSvc *service.AuthService
	Db      data.Queryer
}

func (a *Authenticate) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		token, err := myjwt.JWTFromRequest(req)
		if err != nil && err != http.ErrNoCookie {
			response := myhttp.InvalidRequestErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}

		var user *data.User
		if err == http.ErrNoCookie {
			user, err = data.GetUserCredentialsByLogin(a.Db, "guest")
			if err != nil {
				// guest account has to be there, so create the account if its not
				if err == data.ErrNotFound {
					guest := &data.User{}
					guest.Login.Set("guest")
					guest.Password.Set(xid.New().String())
					if err := guest.PrimaryEmail.Set("guest@rkus.ninja"); err != nil {
						mylog.Log.WithError(err).Error(util.Trace(""))
						response := myhttp.InternalServerErrorResponse("")
						myhttp.WriteResponseTo(rw, response)
						return
					}
					user, err = data.CreateUser(a.Db, guest)
					if err != nil {
						if dErr, ok := err.(data.DataEndUserError); ok {
							if dErr.Code != data.UniqueViolation {
								mylog.Log.WithError(err).Error(util.Trace(""))
								response := myhttp.InternalServerErrorResponse("")
								myhttp.WriteResponseTo(rw, response)
								return
							}
						}
					}
				} else {
					mylog.Log.WithError(err).Error(util.Trace(""))
					response := myhttp.InternalServerErrorResponse("")
					myhttp.WriteResponseTo(rw, response)
					return
				}
			}
		} else {
			payload, err := a.AuthSvc.ValidateJWT(token)
			if err != nil {
				response := myhttp.UnauthorizedErrorResponse(err.Error())
				myhttp.WriteResponseTo(rw, response)
				return
			}

			user, err = data.GetUserCredentials(a.Db, payload.Sub)
			if err != nil {
				response := myhttp.UnauthorizedErrorResponse("user not found")
				myhttp.WriteResponseTo(rw, response)
				http.SetCookie(rw, &http.Cookie{
					Name:     "access_token",
					Value:    "",
					Expires:  time.Unix(0, 0),
					HttpOnly: true,
					// Secure:   true,
				})
				return
			}
		}

		ctx := myctx.NewUserContext(req.Context(), user)
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			response := myhttp.InternalServerErrorResponse("failed to parse requester ip")
			myhttp.WriteResponseTo(rw, response)
			return
		}
		if ip := net.ParseIP(host); ip != nil {
			mask := net.CIDRMask(len(ip)*8, len(ip)*8)
			ipNet := &net.IPNet{IP: ip, Mask: mask}
			ctx = myctx.NewRequesterIpContext(ctx, ipNet)
		} else {
			response := myhttp.InternalServerErrorResponse("failed to parse requester ip")
			myhttp.WriteResponseTo(rw, response)
			return
		}

		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

package middleware

import (
	"context"
	"net"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/service"
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
	Svcs *service.Services
}

func (a *Authenticate) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		token, err := myjwt.JWTFromRequest(req)
		if err != nil {
			response := myhttp.UnauthorizedErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}

		payload, err := a.Svcs.Auth.ValidateJWT(token)
		if err != nil {
			response := myhttp.UnauthorizedErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}

		user, err := a.Svcs.User.Get(payload.Sub)
		if err != nil {
			response := myhttp.UnauthorizedErrorResponse("user not found")
			myhttp.WriteResponseTo(rw, response)
			return
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

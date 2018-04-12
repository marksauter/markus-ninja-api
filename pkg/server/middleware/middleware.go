package middleware

import (
	"context"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type Middleware interface {
	Use(http.Handler) http.Handler
}

type AddContext struct {
	Ctx context.Context
}

func (a *AddContext) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(rw, req.WithContext(a.Ctx))
	})
}

type Authenticate struct {
	AuthSvc  *service.AuthService
	UserRepo *repo.UserRepo
	Logger   *mylog.Logger
}

func (a *Authenticate) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		token, err := myctx.Token.FromRequest(req)
		if err != nil {
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
		}

		payload, err := a.AuthSvc.ValidateJWT(token)
		if err != nil {
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
		}

		user, err := a.UserRepo.Get(payload.Sub)
		if err != nil {
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
		}

		ctx := myctx.User.NewContext(req.Context(), user)
		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

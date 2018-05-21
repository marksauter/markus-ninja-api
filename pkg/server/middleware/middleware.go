package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
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
	Svcs     *service.Services
	UserRepo *repo.UserRepo
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

		a.UserRepo.Open(req.Context())
		_, err = a.UserRepo.AddPermission(perm.Read)
		if err != nil {
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}

		user, err := a.UserRepo.Get(payload.Sub)
		if err != nil {
			response := myhttp.UnauthorizedErrorResponse("user not found")
			myhttp.WriteResponseTo(rw, response)
			return
		}
		a.UserRepo.Close()

		ctx := repo.NewUserContext(req.Context(), user)
		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

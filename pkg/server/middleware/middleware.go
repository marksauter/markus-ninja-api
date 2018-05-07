package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type Middleware interface {
	Use(http.Handler) http.Handler
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
	Svcs  *service.Services
	Repos *repo.Repos
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

		queryPerm, err := a.Repos.Perm().GetQueryPermission(perm.ReadUser, "SELF")
		if err != nil {
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}
		a.Repos.User().AddPermission(queryPerm)

		user, err := a.Repos.User().Get(payload.Sub)
		if err != nil {
			response := myhttp.UnauthorizedErrorResponse("user not found")
			myhttp.WriteResponseTo(rw, response)
			return
		}
		a.Repos.User().ClearPermissions()

		ctx := myctx.User.NewContext(req.Context(), user)
		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

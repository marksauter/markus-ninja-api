package route

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func UserAssets(svcs *service.Services) http.Handler {
	userAssetsHandler := UserAssetsHandler{}
	return middleware.CommonMiddleware.Then(userAssetsHandler)
}

type UserAssetsHandler struct{}

func (h UserAssetsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	routeVars := mux.Vars(req)

	userId := routeVars["user_id"]
	uid, err := oid.NewFromShort("User", userId)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create new oid")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	svc, err := service.NewStorageService()
	if err != nil {
		mylog.Log.WithError(err).Error("failed to start asset service")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	key := routeVars["key"]
	asset, err := svc.Get(uid, key)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to get file")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	defer asset.Close()

	n, err := io.Copy(rw, asset)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to copy asset to response writer")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	mylog.Log.WithField("size", n).Info("successfully streamed file")
	return
}
package route

import (
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
)

func Permissions() http.Handler {
	permissionsHandler := PermissionsHandler{}
	return middleware.CommonMiddleware.Then(permissionsHandler)
}

type PermissionsHandler struct{}

func (h PermissionsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, "static/permissions.html")
}

package route

import (
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
)

func Index() http.Handler {
	indexHandler := IndexHandler{}
	return middleware.CommonMiddleware.Then(indexHandler)
}

type IndexHandler struct{}

func (h IndexHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, "static/index.html")
}

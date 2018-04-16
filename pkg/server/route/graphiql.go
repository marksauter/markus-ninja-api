package route

import (
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
)

func GraphiQL() http.Handler {
	graphiQLHandler := GraphiQLHandler{}
	return middleware.CommonMiddleware.Then(graphiQLHandler)
}

type GraphiQLHandler struct{}

func (h GraphiQLHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, "static/graphiql.html")
}

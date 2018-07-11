package route

import (
	"net/http"
)

type IndexHandler struct{}

func (h IndexHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, "static/index.html")
}

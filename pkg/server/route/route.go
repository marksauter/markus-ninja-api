package route

import (
	"net/http"
)

var Root = RootHandler{}

type RootHandler struct{}

func (h RootHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("hello, world!"))
}

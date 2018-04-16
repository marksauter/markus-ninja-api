package repo

import (
	"net/http"
)

type Middleware struct {
	Repos *Repos
}

func (m *Middleware) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		m.Repos.OpenAll()
		defer m.Repos.CloseAll()
		h.ServeHTTP(rw, req)
	})
}

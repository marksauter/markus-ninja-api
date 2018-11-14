package route

import (
	"net/http"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/cors"
)

type RemoveTokenHandler struct{}

func (h RemoveTokenHandler) Cors() *cors.Cors {
	branch := util.GetRequiredEnv("BRANCH")
	allowedOrigins := []string{"ma.rkus.ninja"}
	if branch != "production" {
		allowedOrigins = append(allowedOrigins, "http://localhost:*")
	}

	return cors.New(cors.Options{
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodOptions, http.MethodGet},
		AllowedOrigins:   allowedOrigins,
		// Debug: true,
	})
}

func (h RemoveTokenHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	if req.Method != http.MethodGet {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	http.SetCookie(rw, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		// Secure:   true,
	})
	return
}

package middleware

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

type Middleware interface {
	Middleware(http.Handler) http.Handler
}

type AddContext struct {
	Ctx context.Context
}

func (a *AddContext) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(rw, req.WithContext(a.Ctx))
	})
}

type AccessLogger struct {
	DebugMode bool
}

func (l *AccessLogger) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		log, ok := myctx.Log.FromContext(ctx)
		if !ok {
			log.Fatal("Log service not in context")
		}
		log.Infof("%s %s %s %s", req.RemoteAddr, req.Method, req.URL, req.Proto)
		log.Infof("User agent : %s", req.UserAgent())
		if l.DebugMode {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Errorf("Reading request body error: %s", err)
			}
			reqStr := ioutil.NopCloser(bytes.NewBuffer(body))
			log.Debugf("Request body: %v", reqStr)
			req.Body = reqStr
		}
		h.ServeHTTP(rw, req)
	})
}

// type UseJwt struct {
//   Jwt *jwt.Jwt
// }
//
// func (u *UseJwt) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
//   var jwtKms *myaws.JwtKms
//   var ok bool
//   if jwtKms, ok := r.Context().Value("JwtKms").(myaws.JwtKms); !ok {
//     log.Fatalf(`"JwtKms" not in request context: got %q`, jwtKms)
//   }
//   var user model.User
//   if user, ok = r.Context().Value("user").(User); !ok {
//     log.Fatalf(`"user" not in request context: got %q`, user)
//   }
//   payload := jwt.NewPayload(&jwt.NewPayloadInput{
//     Id: user.Id,
//   })
//   // token := jwtKms.Encrypt(payload)
//   token := jwt.Token{Payload: payload}
//
//   user.Token = token.String()
//
//   ctx := context.WithValue(r.Context(), "user", user)
//   log.Print(ctx)
//
//   next.ServeHTTP(w, r.WithContext(ctx))
// }

// func (u *SendUserToken) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
//   payload := jwt.NewPayload(&jwt.NewPayloadInput{Id: u.User.Id})
//   token := u.Jwt.Encrypt(&payload)
//   rw.WriteHeader(http.StatusOK)
//   rw.Header().Set("Content-Type", "application/json")
//   io.WriteString(rw, fmt.Sprintf(`{"message": "OK", "token": %s}`, token.String()))
// }

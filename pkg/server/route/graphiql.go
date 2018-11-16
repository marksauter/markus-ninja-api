package route

import (
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/cors"
)

type GraphiQLHandler struct{}

func (h GraphiQLHandler) Cors() *cors.Cors {
	branch := util.GetRequiredEnv("BRANCH")
	allowedOrigins := []string{"ma.rkus.ninja"}
	if branch != "production" {
		allowedOrigins = append(allowedOrigins, "http://localhost:*")
	}

	return cors.New(cors.Options{
		AllowedMethods: []string{http.MethodOptions, http.MethodGet},
		AllowedOrigins: allowedOrigins,
	})
}

func (h GraphiQLHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// viewer, ok := myctx.UserFromContext(req.Context())
	// if !ok {
	//   mylog.Log.Error("viewer not found")
	// }
	// authorized := false
	// for _, r := range viewer.Roles.Elements {
	//   if r.String == data.AdminRole {
	//     authorized = true
	//   }
	// }
	// if !authorized {
	//   response := myhttp.AccessDeniedErrorResponse()
	//   myhttp.WriteResponseTo(rw, response)
	//   return
	// }
	http.ServeFile(rw, req, "static/graphiql.html")
}

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

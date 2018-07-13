package route

import (
	"net/http"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/rs/cors"
)

var GraphQLSchemaCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodGet},
	AllowedOrigins: []string{"ma.rkus.ninja", "http://localhost:*"},
})

type GraphQLSchemaHandler struct {
	Schema *graphql.Schema
}

func (h GraphQLSchemaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	json, err := h.Schema.ToJSON()
	if err != nil {
		errResponse := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, errResponse)
		return
	}
	rw.Write(json)
}

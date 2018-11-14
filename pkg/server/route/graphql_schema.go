package route

import (
	"errors"
	"net/http"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/rs/cors"
)

type GraphQLSchemaHandler struct {
	Conf   *myconf.Config
	Schema *graphql.Schema
}

func (h GraphQLSchemaHandler) Cors() *cors.Cors {
	branch := util.GetRequiredEnv("BRANCH")
	allowedOrigins := []string{"ma.rkus.ninja"}
	if branch != "production" {
		allowedOrigins = append(allowedOrigins, "http://localhost:*")
	}

	return cors.New(cors.Options{
		AllowedHeaders: []string{"Content-Type"},
		AllowedMethods: []string{http.MethodOptions, http.MethodGet},
		AllowedOrigins: allowedOrigins,
	})
}

func (h GraphQLSchemaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.Conf == nil || h.Schema == nil {
		err := errors.New("route inproperly setup")
		mylog.Log.WithError(err).Error(util.Trace(""))
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

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

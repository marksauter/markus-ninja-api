/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/resolver"
	"github.com/marksauter/markus-ninja-api/pkg/schema"
	"github.com/marksauter/markus-ninja-api/pkg/server"
	"github.com/marksauter/markus-ninja-api/pkg/utils"
)

var JwtKms = myaws.NewJwtKms()

func main() {
	r := mux.NewRouter()
	r.Handle("/", server.CommonHandlers.Then(server.RootRouteHandler{}))

	graphqlSchema := graphql.MustParseSchema(
		schema.GetRootSchema(),
		&resolver.Resolver{},
	)

	r.Handle("/graphql", server.CommonHandlers.Then(
		server.GraphQLHandler{Schema: graphqlSchema},
	))

	r.Handle("/graphiql", server.CommonHandlers.ThenFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			http.ServeFile(rw, req, "static/graphiql.html")
		},
	))

	port := utils.GetOptionalEnv("PORT", "5000")
	address := ":" + port
	log.Fatal(http.ListenAndServe(address, r))
}

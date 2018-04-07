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
	"context"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/justinas/alice"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/resolver"
	"github.com/marksauter/markus-ninja-api/pkg/schema"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/server/route"
	"github.com/marksauter/markus-ninja-api/pkg/utils"
)

var JwtKms = myaws.NewJwtKms()

func main() {
	utils.LoadEnv()
	db, err := mydb.Open()
	if err != nil {
		log.Fatalf("Unable to connect to db: %s \n", err)
	}
	defer db.Close()
	ctx := context.Background()
	logger := mylog.NewLogger(true)

	ctx = myctx.Log.NewContext(ctx, logger)

	addContext := middleware.AddContext{Ctx: ctx}
	accessLogger := middleware.AccessLogger{DebugMode: true}
	CommonMiddleware := alice.New(
		addContext.Middleware,
		accessLogger.Middleware,
		handlers.RecoveryHandler(),
	)

	r := mux.NewRouter()

	r.Handle("/", CommonMiddleware.ThenFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			http.ServeFile(rw, req, "static/index.html")
		},
	))

	graphqlSchema := graphql.MustParseSchema(
		schema.GetRootSchema(),
		&resolver.Resolver{},
	)

	r.Handle("/graphql", CommonMiddleware.Then(
		route.GraphQLHandler{Schema: graphqlSchema},
	))

	r.Handle("/graphiql", CommonMiddleware.ThenFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			http.ServeFile(rw, req, "static/graphiql.html")
		},
	))

	r.Handle("/login", CommonMiddleware.Then(route.Login))

	r.Handle("/maria", CommonMiddleware.ThenFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			// Connect and check the server version
			var version string
			err = db.QueryRow("SELECT VERSION()").Scan(&version)
			switch {
			case err != nil:
				log.Fatal(err)
			default:
				fmt.Fprintf(rw, "Connected to: %s", version)
			}
		},
	))
	port := utils.GetOptionalEnv("PORT", "5000")
	address := ":" + port
	log.Fatal(http.ListenAndServe(address, r))
}

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
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/justinas/alice"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/resolver"
	"github.com/marksauter/markus-ninja-api/pkg/schema"
	"github.com/marksauter/markus-ninja-api/pkg/server/route"
	"github.com/marksauter/markus-ninja-api/pkg/service"
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

	logger := mylog.NewLogger(true)
	userRepo := repo.NewUserRepo(service.NewUserService(db, logger))

	CommonMiddleware := alice.New(
		logger.AccessMiddleware,
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
		&resolver.Resolver{
			UserRepo: userRepo,
		},
	)

	r.Handle("/graphql", CommonMiddleware.Then(
		route.GraphQLHandler{Schema: graphqlSchema},
	))

	r.Handle("/graphiql", CommonMiddleware.ThenFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			http.ServeFile(rw, req, "static/graphiql.html")
		},
	))

	r.Handle("/login", CommonMiddleware.Then(
		route.LoginHandler{UserRepo: userRepo},
	))

	r.Handle("/db", CommonMiddleware.ThenFunc(
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

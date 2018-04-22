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
	"net/http"

	"github.com/gorilla/mux"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/resolver"
	"github.com/marksauter/markus-ninja-api/pkg/schema"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/server/route"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func main() {
	util.LoadEnv()
	db, err := mydb.Open()
	if err != nil {
		mylog.Log.WithField("error", err).Fatal("unable to connect to database")
	}
	defer db.Close()

	if err = initDB(db); err != nil {
		mylog.Log.WithField("error", err).Fatal("error initializing database")
	}

	svcs := service.NewServices(db)
	repos := repo.NewRepos(svcs)
	graphQLSchema := graphql.MustParseSchema(
		schema.GetRootSchema(),
		&resolver.RootResolver{
			Repos: repos,
		},
	)

	r := mux.NewRouter()

	r.Handle("/", route.Index())
	r.Handle("/graphql", route.GraphQL(graphQLSchema, svcs.Auth, repos))
	r.Handle("/graphiql", route.GraphiQL())
	r.Handle("/token", route.Token(svcs.Auth, repos.User()))

	r.Handle("/db", middleware.CommonMiddleware.ThenFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			// Connect and check the server version
			var version string
			err = db.QueryRow("SELECT VERSION()").Scan(&version)
			if err != nil {
				mylog.Log.Fatal(err)
				return
			}
			fmt.Fprintf(rw, "Connected to: %s", version)
		},
	))
	port := util.GetOptionalEnv("PORT", "5000")
	address := ":" + port
	mylog.Log.Fatal(http.ListenAndServe(address, r))
}

func initDB(db *mydb.DB) error {
	defer func() {
		if r := recover(); r != nil {
			mylog.Log.Debug(r)
		}
	}()
	svcs := service.NewServices(db)

	roleNames := []string{"ADMIN", "MEMBER", "USER"}

	for _, name := range roleNames {
		if _, err := svcs.Role.Create(name); err != nil {
			mylog.Log.WithError(err).Fatal("error during role creation")
			return err
		}
	}

	modelTypes := []interface{}{
		new(model.User),
	}

	for _, model := range modelTypes {
		if err := svcs.Perm.CreatePermissionSuite(model); err != nil {
			mylog.Log.WithError(err).Fatal("error during permission suite creation")
			return err
		}
	}

	return nil
}

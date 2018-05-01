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
	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/resolver"
	"github.com/marksauter/markus-ninja-api/pkg/schema"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/server/route"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

func main() {
	config := myconf.Load("config")
	dbConfig := pgx.ConnConfig{
		User:     config.DBUser,
		Password: config.DBPassword,
		Host:     config.DBHost,
		Port:     config.DBPort,
		Database: config.DBName,
	}
	db, err := mydb.Open(dbConfig)
	if err != nil {
		mylog.Log.WithField("error", err).Fatal("unable to connect to database")
	}
	defer db.Close()

	if err = initDB(db); err != nil {
		mylog.Log.WithField("error", err).Fatal("error initializing database")
	}

	svcs := data.NewServices(db)
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
	r.Handle("/permissions", route.Permissions())
	r.Handle("/signup", route.Signup(svcs.Auth, repos))
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
	svcs := data.NewServices(db)

	roleNames := []string{"ADMIN", "MEMBER", "SELF", "USER"}

	for _, name := range roleNames {
		if _, err := svcs.Role.Create(name); err != nil {
			mylog.Log.WithError(err).Fatal("error during role creation")
			return err
		}
	}

	modelTypes := []interface{}{
		new(data.UserModel),
	}

	for _, model := range modelTypes {
		if err := svcs.Perm.CreatePermissionSuite(model); err != nil {
			mylog.Log.WithError(err).Fatal("error during permission suite creation")
			return err
		}
	}

	publicReadUserFields := []string{
		"bio",
		"email",
		"id",
		"login",
		"name",
	}
	err := svcs.Perm.UpdateOperationForFields(
		perm.ReadUser,
		publicReadUserFields,
		perm.Everyone,
	)
	if err != nil {
		mylog.Log.WithError(err).Fatal("error during permission update")
		return err
	}

	publicCreateUserFields := []string{
		"primary_email",
		"id",
		"login",
	}
	err = svcs.Perm.UpdateOperationForFields(
		perm.CreateUser,
		publicCreateUserFields,
		perm.Everyone,
	)
	if err != nil {
		mylog.Log.WithError(err).Fatal("error during permission update")
		return err
	}

	adminPermissionsSQL := `
		SELECT
			r.id role_id,
			p.id permission_id
		FROM
			role r
		INNER JOIN permission p ON true
		WHERE r.name = 'ADMIN'
	`
	rows, err := db.Query(adminPermissionsSQL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("Unexpected: no permissions found")
		}
		return err
	}
	adminPermissionsCount, err := db.CopyFrom(
		pgx.Identifier{"role_permission"},
		[]string{"role_id", "permission_id"},
		rows,
	)
	if pgErr, ok := err.(pgx.PgError); ok {
		if data.PSQLError(pgErr.Code) != data.UniqueViolation {
			return err
		}
	}
	mylog.Log.WithFields(logrus.Fields{
		"n": adminPermissionsCount,
	}).Infof("role permissions created for ADMIN")

	selfPermissionsSQL := `
		SELECT
			r.id role_id,
			p.id permission_id
		FROM
			role r
		INNER JOIN permission p ON p.type = 'User'
		WHERE r.name = 'SELF'
	`
	rows, err = db.Query(selfPermissionsSQL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("Unexpected: no permissions found")
		}
		return err
	}
	selfPermissionsCount, err := db.CopyFrom(
		pgx.Identifier{"role_permission"},
		[]string{"role_id", "permission_id"},
		rows,
	)
	if pgErr, ok := err.(pgx.PgError); ok {
		if data.PSQLError(pgErr.Code) != data.UniqueViolation {
			return err
		}
	}
	mylog.Log.WithFields(logrus.Fields{
		"n": selfPermissionsCount,
	}).Infof("role permissions created for SELF")

	userPermissionsSQL := `
		SELECT
			r.id role_id,
			p.id permission_id
		FROM
			role r
		INNER JOIN permission p ON p.audience = 'EVERYONE'
		WHERE r.name = 'USER'
	`
	rows, err = db.Query(userPermissionsSQL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("Unexpected: no permissions found")
		}
		return err
	}
	userPermissionsCount, err := db.CopyFrom(
		pgx.Identifier{"role_permission"},
		[]string{"role_id", "permission_id"},
		rows,
	)
	if pgErr, ok := err.(pgx.PgError); ok {
		if data.PSQLError(pgErr.Code) != data.UniqueViolation {
			return err
		}
	}
	mylog.Log.WithFields(logrus.Fields{
		"n": userPermissionsCount,
	}).Infof("role permissions created for USER")
	return nil
}

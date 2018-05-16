/* Copyright 2016 The Kubernetes Authors.

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
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/resolver"
	"github.com/marksauter/markus-ninja-api/pkg/schema"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/server/route"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

func main() {
	conf := myconf.Load("config.development")
	dbConfig := pgx.ConnConfig{
		User:     conf.DBUser,
		Password: conf.DBPassword,
		Host:     conf.DBHost,
		Port:     conf.DBPort,
		Database: conf.DBName,
	}
	db, err := mydb.Open(dbConfig)
	if err != nil {
		mylog.Log.WithField("error", err).Fatal("unable to connect to database")
	}
	defer db.Close()

	svcs := service.NewServices(conf, db)

	if err = initDB(svcs, db); err != nil {
		mylog.Log.WithField("error", err).Fatal("error initializing database")
	}

	repos := repo.NewRepos(svcs)
	graphQLSchema := graphql.MustParseSchema(
		schema.GetRootSchema(),
		&resolver.RootResolver{
			Repos: repos,
			Svcs:  svcs,
		},
	)

	r := mux.NewRouter()

	r.Handle("/", route.Index())
	r.Handle("/graphql", route.GraphQL(graphQLSchema, svcs, repos))
	r.Handle("/graphiql", route.GraphiQL())
	r.Handle("/permissions", route.Permissions())
	r.Handle("/signup", route.Signup(svcs, repos))
	r.Handle("/token", route.Token(svcs, repos))
	r.Handle("/upload", route.Upload())
	r.Handle("/upload/assets", route.UploadAssets())
	r.Handle(
		"/users/{login}/emails/confirm_verification/{token}",
		route.ConfirmVerification(svcs),
	)
	r.Handle(
		"/users/{login}/emails/request_verification",
		route.RequestEmailVerification(svcs),
	)
	r.Handle(
		"/users/{login}/request_password_reset",
		route.RequestPasswordReset(svcs),
	)
	r.Handle("/users/{login}/reset_password", route.ResetPassword(svcs))

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

func initDB(svcs *service.Services, db *mydb.DB) error {
	defer func() {
		if r := recover(); r != nil {
			mylog.Log.Panic(r)
		}
	}()

	roleNames := []string{"ADMIN", "MEMBER", "SELF", "USER"}

	for _, name := range roleNames {
		if _, err := svcs.Role.Create(name); err != nil {
			mylog.Log.WithError(err).Fatal("error during role creation")
			return err
		}
	}

	modelTypes := []interface{}{
		new(data.Lesson),
		new(data.Study),
		new(data.User),
	}

	for _, model := range modelTypes {
		if err := svcs.Perm.CreatePermissionSuite(model); err != nil {
			mylog.Log.WithError(err).Fatal("error during permission suite creation")
			return err
		}
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

	guestId, _ := oid.New("User")
	guestPassword, _ := passwd.New("guest")
	guest := &data.User{}
	guest.Id.Set(guestId.String())
	guest.Login.Set("guest")
	guest.Password.Set(guestPassword.Hash())
	guest.PrimaryEmail.Set("guest@rkus.ninja")
	if err := svcs.User.Create(guest); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal("failed to create guest account")
				return err
			}
			mylog.Log.Info("guest account already exists")
		}
	}

	markusId, _ := oid.New("User")
	markusPassword, _ := passwd.New("fender917")
	markus := &data.User{}
	markus.Id.Set(markusId.String())
	markus.Login.Set("markus")
	markus.Password.Set(markusPassword.Hash())
	markus.PrimaryEmail.Set("m@rkus.ninja")
	if err := svcs.User.Create(markus); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal("failed to create markus account")
				return err
			}
			mylog.Log.Info("markus account already exists")
		}
	}
	if err := svcs.Role.GrantUser(markus.Id.String, data.AdminRole); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal("failed to grant markus admin role")
				return err
			}
			mylog.Log.Info("markus is already an admin")
		}
	}

	mylog.Log.Info("database initialized")
	return nil
}

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
	branch := util.GetRequiredEnv("BRANCH")
	confFilename := fmt.Sprintf("config.%s", branch)
	conf := myconf.Load(confFilename)

	if err := initDB(conf); err != nil {
		mylog.Log.WithField("error", err).Fatal("error initializing database")
	}

	dbConfig := pgx.ConnConfig{
		User: "client",
		// Password: conf.DBPassword,
		Host:     conf.DBHost,
		Port:     conf.DBPort,
		Database: conf.DBName,
	}
	db, err := mydb.Open(dbConfig)
	if err != nil {
		mylog.Log.WithField("error", err).Fatal("unable to connect to database")
	}
	defer db.Close()

	svcs, err := service.NewServices(conf, db)
	if err != nil {
		panic(err)
	}

	repos := repo.NewRepos(svcs, db, conf)
	graphQLSchema := graphql.MustParseSchema(
		schema.GetRootSchema(),
		&resolver.RootResolver{
			Db:    db,
			Repos: repos,
			Svcs:  svcs,
		},
	)

	go startRefreshMV(svcs)

	r := mux.NewRouter()

	r.Handle("/", route.Index())
	r.Handle("/graphql", route.GraphQL(graphQLSchema, svcs, repos))
	r.Handle("/graphiql", route.GraphiQL())
	r.Handle("/signup", route.Signup(svcs))
	r.Handle("/token", route.Token(svcs))
	r.Handle("/upload", route.Upload())
	r.Handle("/upload/assets", route.UploadAssets(svcs, repos))
	r.Handle("/user/{login}/emails/{id}/confirm_verification/{token}",
		route.ConfirmVerification(svcs, db),
	)
	r.Handle("/user/assets/{user_id}/{key}",
		route.UserAssets(svcs),
	)

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

func initDB(conf *myconf.Config) error {
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

	svcs, err := service.NewServices(conf, db)
	if err != nil {
		panic(err)
	}
	err = data.Initialize(db)
	if err != nil {
		return err
	}

	modelTypes := []interface{}{
		new(data.Appled),
		new(data.Email),
		new(data.Enrolled),
		new(data.Event),
		new(data.EVT),
		new(data.Label),
		new(data.Labeled),
		new(data.Lesson),
		new(data.LessonComment),
		new(data.Notification),
		new(data.PRT),
		new(data.Study),
		new(data.Topic),
		new(data.Topiced),
		new(data.User),
		new(data.UserAsset),
	}

	for _, model := range modelTypes {
		if err := svcs.Perm.CreatePermissionSuite(model); err != nil {
			mylog.Log.WithError(err).Fatal("error during permission suite creation")
			return err
		}
	}

	permissions, err := myconf.LoadPermissions()
	if err != nil {
		panic(err)
	}

	for _, p := range permissions.Permissions {
		if !p.Authenticated {
			err := svcs.Perm.UpdateOperationAudience(&p.Operation, mytype.Everyone, p.Fields)
			if err != nil {
				return err
			}
		} else {
			err := svcs.Perm.ConnectRoles(&p.Operation, p.Fields, p.Roles)
			if err != nil {
				return err
			}
		}
	}

	adminPermissionsSQL := `
		INSERT INTO role_permission(permission_id, role) (
			SELECT
				permission.id,
				role.name
			FROM role
			JOIN permission ON true
			WHERE role.name = 'ADMIN'
		) ON CONFLICT ON CONSTRAINT role_permission_pkey DO NOTHING
	`
	commandTag, err := db.Exec(adminPermissionsSQL)
	if err != nil {
		return err
	}
	mylog.Log.WithFields(logrus.Fields{
		"n": commandTag.RowsAffected(),
	}).Infof("role permissions created for ADMIN")

	userPermissionsSQL := `
		INSERT INTO role_permission(permission_id, role) (
			SELECT
				permission.id,
				role.name
			FROM role
			JOIN permission ON permission.audience = 'EVERYONE'
			WHERE role.name = 'USER'
		) ON CONFLICT ON CONSTRAINT role_permission_pkey DO NOTHING
	`
	commandTag, err = db.Exec(userPermissionsSQL)
	if err != nil {
		return err
	}
	mylog.Log.WithFields(logrus.Fields{
		"n": commandTag.RowsAffected(),
	}).Infof("role permissions created for USER")

	guestId, _ := mytype.NewOID("User")
	guest := &data.User{}
	guest.Id.Set(guestId)
	guest.Login.Set("guest")
	guest.Password.Set("guest")
	if err := guest.PrimaryEmail.Set("guest@rkus.ninja"); err != nil {
		return err
	}
	if _, err := svcs.User.Create(guest); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal("failed to create guest account")
				return err
			}
			mylog.Log.Info("guest account already exists")
		} else {
			return err
		}
	}

	markusId, _ := mytype.NewOID("User")
	markus := &data.User{}
	markus.Id.Set(markusId)
	markus.Login.Set("markus")
	markus.Password.Set("fender917")
	if err := markus.PrimaryEmail.Set("m@rkus.ninja"); err != nil {
		return err
	}
	if _, err := svcs.User.Create(markus); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal("failed to create markus account")
				return err
			}
			mylog.Log.Info("markus account already exists")
			markus, err = svcs.User.GetByLogin("markus")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	markusIsAdmin := false
	for _, r := range markus.Roles.Elements {
		if r.String == data.AdminRole {
			markusIsAdmin = true
		}
	}
	if !markusIsAdmin {
		if err := svcs.Role.GrantUser(markus.Id.String, data.AdminRole); err != nil {
			if dfErr, ok := err.(data.DataFieldError); ok {
				if dfErr.Code != data.DuplicateField {
					mylog.Log.WithError(err).Fatal("failed to grant markus admin role")
					return err
				}
				mylog.Log.Info("markus is already an admin")
			} else {
				return err
			}
		}
	}

	testUserId, _ := mytype.NewOID("User")
	testUser := &data.User{}
	testUser.Id.Set(testUserId)
	testUser.Login.Set("test")
	testUser.Password.Set("test")
	if err := testUser.PrimaryEmail.Set("test@example.com"); err != nil {
		return err
	}
	if _, err := svcs.User.Create(testUser); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal("failed to create testUser account")
				return err
			}
			mylog.Log.Info("testUser account already exists")
			testUser, err = svcs.User.GetByLogin("test")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	testUserIsUser := false
	for _, r := range testUser.Roles.Elements {
		if r.String == data.UserRole {
			testUserIsUser = true
		}
	}
	if !testUserIsUser {
		if err := svcs.Role.GrantUser(testUser.Id.String, data.UserRole); err != nil {
			if dfErr, ok := err.(data.DataFieldError); ok {
				if dfErr.Code != data.DuplicateField {
					mylog.Log.WithError(err).Fatal("failed to grant testUser admin role")
					return err
				}
				mylog.Log.Info("testUser is already an admin")
			} else {
				return err
			}
		}
	}

	mylog.Log.Info("database initialized")
	return nil
}

func startRefreshMV(svcs *service.Services) {
	for {
		go svcs.User.RefreshSearchIndex()
		time.Sleep(10 * time.Second)
		go svcs.Study.RefreshSearchIndex()
		time.Sleep(10 * time.Second)
		go svcs.Lesson.RefreshSearchIndex()
		time.Sleep(10 * time.Second)
		go svcs.Label.RefreshSearchIndex()
		time.Sleep(10 * time.Minute)
		go svcs.Topic.RefreshSearchIndex()
		time.Sleep(30 * time.Minute)
	}
}

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

	svcs, err := service.NewServices(conf)
	if err != nil {
		mylog.Log.WithField("error", err).Fatal("unable to start services")
	}

	repos := repo.NewRepos(db, svcs)
	schema := graphql.MustParseSchema(
		schema.GetRootSchema(),
		&resolver.RootResolver{
			Repos: repos,
			Svcs:  svcs,
		},
	)

	go startRefreshMV(db)

	r := mux.NewRouter()

	authMiddleware := middleware.Authenticate{Db: db, AuthSvc: svcs.Auth}

	graphQLHandler := route.GraphQLHandler{Schema: schema, Repos: repos}
	graphQLSchemaHandler := route.GraphQLSchemaHandler{Schema: schema}
	graphiQLHandler := route.GraphiQLHandler{}
	confirmVerificationHandler := route.ConfirmVerificationHandler{db}
	indexHandler := route.IndexHandler{}
	tokenHandler := route.TokenHandler{svcs.Auth, db}
	signupHandler := route.SignupHandler{svcs.Auth, db}
	uploadHandler := route.UploadHandler{}
	uploadAssetsHandler := route.UploadAssetsHandler{repos}
	userAssetsHandler := route.UserAssetsHandler{svcs.Storage}

	graphql := middleware.CommonMiddleware.Append(
		route.GraphQLCors.Handler,
		authMiddleware.Use,
		repos.Use,
	).Then(graphQLHandler)
	graphQLSchema := middleware.CommonMiddleware.Append(
		route.GraphQLSchemaCors.Handler,
	).Then(graphQLSchemaHandler)
	graphiql := middleware.CommonMiddleware.Append(
		route.GraphiQLCors.Handler,
	).Then(graphiQLHandler)
	confirmVerification := middleware.CommonMiddleware.Append(
		route.ConfirmVerificationCors.Handler,
	).Then(confirmVerificationHandler)
	index := middleware.CommonMiddleware.Then(indexHandler)
	token := middleware.CommonMiddleware.Append(
		route.TokenCors.Handler,
	).Then(tokenHandler)
	signup := middleware.CommonMiddleware.Append(
		route.SignupCors.Handler,
		authMiddleware.Use,
	).Then(signupHandler)
	upload := middleware.CommonMiddleware.Then(uploadHandler)
	uploadAssets := middleware.CommonMiddleware.Append(
		route.UploadAssetsCors.Handler,
		authMiddleware.Use,
		repos.Use,
	).Then(uploadAssetsHandler)
	userAssets := middleware.CommonMiddleware.Append(
		route.UserAssetsCors.Handler,
	).Then(userAssetsHandler)

	r.Handle("/", index)
	r.Handle("/graphql", graphql)
	r.Handle("/graphql/schema", graphQLSchema)
	r.Handle("/graphiql", graphiql)
	r.Handle("/signup", signup)
	r.Handle("/token", token)
	r.Handle("/upload", upload)
	r.Handle("/upload/assets", uploadAssets)
	r.Handle("/user/{login}/emails/{id}/confirm_verification/{token}",
		confirmVerification,
	)
	r.Handle("/user/assets/{user_id}/{key}", userAssets)

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
		if err := data.CreatePermissionSuite(db, model); err != nil {
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
			err := data.UpdatePermissionAudience(db, &p.Operation, mytype.Everyone, p.Fields)
			if err != nil {
				return err
			}
		} else {
			err := data.ConnectRolePermissions(db, &p.Operation, p.Fields, p.Roles)
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
	if _, err := data.CreateUser(db, guest); err != nil {
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
	if _, err := data.CreateUser(db, markus); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal("failed to create markus account")
				return err
			}
			mylog.Log.Info("markus account already exists")
			markus, err = data.GetUserByLogin(db, "markus")
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
		if err := data.GrantUserRoles(db, markus.Id.String, data.AdminRole); err != nil {
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
	if _, err := data.CreateUser(db, testUser); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal("failed to create testUser account")
				return err
			}
			mylog.Log.Info("testUser account already exists")
			testUser, err = data.GetUserByLogin(db, "test")
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
		if err := data.GrantUserRoles(db, testUser.Id.String, data.UserRole); err != nil {
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

func startRefreshMV(db data.Queryer) {
	for {
		go data.RefreshUserSearchIndex(db)
		time.Sleep(10 * time.Second)
		go data.RefreshStudySearchIndex(db)
		time.Sleep(10 * time.Second)
		go data.RefreshLessonSearchIndex(db)
		time.Sleep(10 * time.Second)
		go data.RefreshTopicSearchIndex(db)
		time.Sleep(10 * time.Second)
		go data.RefreshLabelSearchIndex(db)
		time.Sleep(10 * time.Second)
		go data.RefreshUserAssetSearchIndex(db)
		time.Sleep(30 * time.Minute)
	}
}

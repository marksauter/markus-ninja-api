package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
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
	"github.com/rs/xid"
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

	repos := repo.NewRepos(db, conf)
	schema := graphql.MustParseSchema(
		schema.GetRootSchema(),
		&resolver.RootResolver{
			Conf:  conf,
			Repos: repos,
			Svcs:  svcs,
		},
	)

	r := mux.NewRouter()
	r.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	authMiddleware := middleware.Authenticate{Db: db, AuthSvc: svcs.Auth}

	graphQLHandler := route.GraphQLHandler{Conf: conf, Schema: schema, Repos: repos}
	graphQLSchemaHandler := route.GraphQLSchemaHandler{Conf: conf, Schema: schema}
	graphiQLHandler := route.GraphiQLHandler{}
	confirmVerificationHandler := route.ConfirmVerificationHandler{Conf: conf, Db: db}
	indexHandler := route.IndexHandler{}
	previewHandler := route.PreviewHandler{Conf: conf, Repos: repos}
	tokenHandler := route.TokenHandler{AuthSvc: svcs.Auth, Conf: conf, Db: db}
	removeTokenHandler := route.RemoveTokenHandler{}
	signupHandler := route.SignupHandler{AuthSvc: svcs.Auth, Conf: conf, Db: db}
	uploadHandler := route.UploadHandler{}
	uploadAssetsHandler := route.UploadAssetsHandler{Conf: conf, Repos: repos, StorageSvc: svcs.Storage}
	userAssetsHandler := route.UserAssetsHandler{Conf: conf, Repos: repos, StorageSvc: svcs.Storage}

	graphql := middleware.CommonMiddleware.Append(
		graphQLHandler.Cors().Handler,
		authMiddleware.Use,
		repos.Use,
	).Then(graphQLHandler)
	graphQLSchema := middleware.CommonMiddleware.Append(
		graphQLSchemaHandler.Cors().Handler,
	).Then(graphQLSchemaHandler)
	graphiql := middleware.CommonMiddleware.Append(
		graphiQLHandler.Cors().Handler,
	).Then(graphiQLHandler)
	confirmVerification := middleware.CommonMiddleware.Append(
		confirmVerificationHandler.Cors().Handler,
	).Then(confirmVerificationHandler)
	index := middleware.CommonMiddleware.Then(indexHandler)
	preview := middleware.CommonMiddleware.Append(
		previewHandler.Cors().Handler,
		authMiddleware.Use,
		repos.Use,
	).Then(previewHandler)
	token := middleware.CommonMiddleware.Append(
		tokenHandler.Cors().Handler,
	).Then(tokenHandler)
	removeToken := middleware.CommonMiddleware.Append(
		removeTokenHandler.Cors().Handler,
	).Then(removeTokenHandler)
	signup := middleware.CommonMiddleware.Append(
		signupHandler.Cors().Handler,
		authMiddleware.Use,
	).Then(signupHandler)
	upload := middleware.CommonMiddleware.Then(uploadHandler)
	uploadAssets := middleware.CommonMiddleware.Append(
		uploadAssetsHandler.Cors().Handler,
		authMiddleware.Use,
		repos.Use,
	).Then(uploadAssetsHandler)
	userAssets := middleware.CommonMiddleware.Append(
		userAssetsHandler.Cors().Handler,
		authMiddleware.Use,
		repos.Use,
	).Then(userAssetsHandler)

	r.Handle("/", index)
	r.Handle("/graphql", graphql)
	r.Handle("/graphql/schema", graphQLSchema)
	r.Handle("/graphiql", graphiql)
	r.Handle("/preview", preview)
	r.Handle("/signup", signup)
	r.Handle("/token", token)
	r.Handle("/remove_token", removeToken)
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

	router := http.TimeoutHandler(r, 5*time.Second, "Timeout!")

	port := util.GetOptionalEnv("PORT", "5000")
	address := ":" + port
	mylog.Log.Fatal(http.ListenAndServe(address, router))
}

func initDB(conf *myconf.Config) error {
	branch := util.GetRequiredEnv("BRANCH")

	dbConfig := pgx.ConnConfig{
		User:     conf.DBUser,
		Password: conf.DBPassword,
		Host:     conf.DBHost,
		Port:     conf.DBPort,
		Database: conf.DBName,
	}
	db, err := mydb.Open(dbConfig)
	if err != nil {
		mylog.Log.WithError(err).Fatal(util.Trace("unable to connect to database"))
	}
	defer db.Close()

	err = data.Initialize(db)
	if err != nil {
		return err
	}

	modelTypes := []interface{}{
		new(data.Appled),
		new(data.Asset),
		new(data.Course),
		new(data.CourseLesson),
		new(data.Email),
		new(data.EVT),
		new(data.Enrolled),
		new(data.Event),
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
			mylog.Log.WithError(err).Fatal(util.Trace("error during permission suite creation"))
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

	guestID, _ := mytype.NewOID("User")
	guest := &data.User{}
	guest.ID.Set(guestID)
	guest.Login.Set("guest")
	guest.Password.Set(xid.New().String())
	if err := guest.PrimaryEmail.Set("guest@rkus.ninja"); err != nil {
		return err
	}
	if _, err := data.CreateUser(db, guest); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal(util.Trace("failed to create guest account"))
				return err
			}
		} else {
			return err
		}
	}

	markus := &data.User{}
	markus.Login.Set("markus")
	markus.Password.Set(xid.New().String())
	if err := markus.PrimaryEmail.Set("m@rkus.ninja"); err != nil {
		return err
	}
	if _, err := data.CreateUser(db, markus); err != nil {
		if dfErr, ok := err.(data.DataFieldError); ok {
			if dfErr.Code != data.DuplicateField {
				mylog.Log.WithError(err).Fatal(util.Trace("failed to create markus account"))
				return err
			}
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
		if err := data.GrantUserRoles(db, markus.ID.String, data.AdminRole); err != nil {
			if dfErr, ok := err.(data.DataFieldError); ok {
				if dfErr.Code != data.DuplicateField {
					mylog.Log.WithError(err).Fatal(util.Trace("failed to grant markus admin role"))
					return err
				}
			} else {
				return err
			}
		}
	}

	if branch == "test" {
		testUser := &data.User{}
		testUser.Login.Set("test")
		testUser.Password.Set("test")
		if err := testUser.PrimaryEmail.Set("test@example.com"); err != nil {
			return err
		}
		if _, err := data.CreateUser(db, testUser); err != nil {
			if dfErr, ok := err.(data.DataFieldError); ok {
				if dfErr.Code != data.DuplicateField {
					mylog.Log.WithError(err).Fatal(util.Trace("failed to create testUser account"))
					return err
				}
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
			if err := data.GrantUserRoles(db, testUser.ID.String, data.UserRole); err != nil {
				if dfErr, ok := err.(data.DataFieldError); ok {
					if dfErr.Code != data.DuplicateField {
						mylog.Log.WithError(err).Fatal(util.Trace("failed to grant testUser user role"))
						return err
					}
				} else {
					return err
				}
			}
		}
	}

	mylog.Log.Info("database initialized")
	return nil
}

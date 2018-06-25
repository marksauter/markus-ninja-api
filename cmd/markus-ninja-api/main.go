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

	if err := initDB(svcs, db); err != nil {
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
		route.ConfirmVerification(svcs),
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

func initDB(svcs *service.Services, db *mydb.DB) error {
	err := data.Initialize(db)
	if err != nil {
		return err
	}

	modelTypes := []interface{}{
		new(data.Email),
		new(data.Event),
		new(data.EVT),
		new(data.Lesson),
		new(data.LessonComment),
		new(data.PRT),
		new(data.Study),
		new(data.StudyApple),
		new(data.StudyEnroll),
		new(data.Topic),
		new(data.User),
		new(data.UserAsset),
		new(data.UserEnroll),
	}

	for _, model := range modelTypes {
		if err := svcs.Perm.CreatePermissionSuite(model); err != nil {
			mylog.Log.WithError(err).Fatal("error during permission suite creation")
			return err
		}
	}

	adminPermissionsSQL := `
		SELECT
			r.name,
			p.id permission_id
		FROM
			role r
		JOIN permission p ON true
		WHERE r.name = ANY('{"ADMIN", "OWNER"}')
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
		[]string{"role", "permission_id"},
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

	userPermissionsSQL := `
		SELECT
			r.name,
			p.id permission_id
		FROM
			role r
		JOIN permission p ON p.audience = 'EVERYONE'
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
		[]string{"role", "permission_id"},
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
		if r.String == data.AdminRole {
			testUserIsUser = true
		}
	}
	if !testUserIsUser {
		if err := svcs.Role.GrantUser(testUser.Id.String, data.AdminRole); err != nil {
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
		go svcs.Topic.RefreshSearchIndex()
		time.Sleep(30 * time.Minute)
	}
}

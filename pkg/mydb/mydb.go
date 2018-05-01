package mydb

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type DB struct {
	*pgx.ConnPool
}

func Open(config pgx.ConnConfig) (*DB, error) {
	mylog.Log.Println("Connecting to database...")
	pgxConnPoolConfig := pgx.ConnPoolConfig{
		ConnConfig:     config,
		MaxConnections: 3,
		AfterConnect:   nil,
	}
	conn, err := pgx.NewConnPool(pgxConnPoolConfig)
	if err != nil {
		return nil, err
	}

	mylog.Log.Println("Database connected")
	return &DB{conn}, nil
}

type TestDB struct {
	DB *DB
	T  testing.TB
}

func NewTestDB(t testing.TB) *TestDB {
	port, err := strconv.ParseUint(util.GetRequiredEnv("TEST_RDS_PORT"), 10, 16)
	if err != nil {
		t.Fatal(err)
	}
	pgxConfig := pgx.ConnConfig{
		User:     util.GetRequiredEnv("TEST_RDS_USERNAME"),
		Password: util.GetRequiredEnv("TEST_RDS_PASSWORD"),
		Host:     util.GetRequiredEnv("TEST_RDS_HOSTNAME"),
		Port:     uint16(port),
		Database: util.GetRequiredEnv("TEST_RDS_DB_NAME"),
		// Logger:   mylog.Log,
		// LogLevel: pgx.LogLevelDebug,
	}
	db, err := Open(pgxConfig)
	if err != nil {
		t.Fatal(err)
	}
	return &TestDB{DB: db, T: t}
}

func (db *TestDB) Close() {
	db.DB.Close()
}

func (db *TestDB) Empty() error {
	tables := []string{
		"account",
		"role",
		"account_role",
		"permission",
		"role_permission",
	}
	for _, table := range tables {
		_, err := db.DB.Exec(fmt.Sprintf("delete from %s", table))
		if err != nil {
			return err
		}
	}
	return nil
}

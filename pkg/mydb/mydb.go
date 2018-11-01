package mydb

import (
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

type DB struct {
	*pgx.ConnPool
}

func Open(config pgx.ConnConfig) (*DB, error) {
	mylog.Log.Println("Connecting to database...")
	pgxConnPoolConfig := pgx.ConnPoolConfig{
		ConnConfig:     config,
		MaxConnections: 5,
		AfterConnect:   nil,
		AcquireTimeout: 5 * time.Second,
	}
	conn, err := pgx.NewConnPool(pgxConnPoolConfig)
	if err != nil {
		return nil, err
	}

	mylog.Log.Println("Database connected")
	return &DB{conn}, nil
}

var SharedTestDB *TestDB

type TestDB struct {
	DB *DB
	T  testing.TB
}

func NewTestDB(t testing.TB) *TestDB {
	if SharedTestDB == nil {
		config := myconf.Load("config.test")
		dbConfig := pgx.ConnConfig{
			User:     config.DBUser,
			Password: config.DBPassword,
			Host:     config.DBHost,
			Port:     config.DBPort,
			Database: config.DBName,
		}
		db, err := Open(dbConfig)
		if err != nil {
			t.Fatal(err)
		}

		SharedTestDB = &TestDB{DB: db, T: t}
	}

	err := SharedTestDB.Empty()
	if err != nil {
		t.Fatal(err)
	}

	return SharedTestDB
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

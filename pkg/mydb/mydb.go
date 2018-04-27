package mydb

import (
	"strconv"

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
	*DB
}

func NewTestDB() *TestDB {
	port, err := strconv.ParseUint(util.GetRequiredEnv("TEST_RDS_PORT"), 10, 16)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	return &TestDB{db}
}

func (db *TestDB) Close() {
	db.DB.Close()
}

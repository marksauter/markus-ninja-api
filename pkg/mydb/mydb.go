package mydb

import (
	"strconv"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type DB = pgx.ConnPool

func Open() (*DB, error) {
	mylog.Log.Println("Connecting to database...")
	port, err := strconv.ParseUint(util.GetRequiredEnv("RDS_PORT"), 10, 16)
	if err != nil {
		panic(err)
	}
	pgxConfig := pgx.ConnConfig{
		User:     util.GetRequiredEnv("RDS_USERNAME"),
		Password: util.GetRequiredEnv("RDS_PASSWORD"),
		Host:     util.GetRequiredEnv("RDS_HOSTNAME"),
		Port:     uint16(port),
		Database: util.GetRequiredEnv("RDS_DB_NAME"),
		Logger:   mylog.Log,
		LogLevel: pgx.LogLevelDebug,
	}
	pgxConnPoolConfig := pgx.ConnPoolConfig{
		ConnConfig:     pgxConfig,
		MaxConnections: 3,
		AfterConnect:   nil,
	}
	conn, err := pgx.NewConnPool(pgxConnPoolConfig)
	if err != nil {
		return nil, err
	}

	mylog.Log.Println("Database connected")
	return conn, nil
}

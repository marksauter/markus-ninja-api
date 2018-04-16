package mydb

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type DB = sqlx.DB

func Open() (*DB, error) {
	username := util.GetRequiredEnv("RDS_USERNAME")
	password := util.GetRequiredEnv("RDS_PASSWORD")
	hostname := util.GetRequiredEnv("RDS_HOSTNAME")
	port := util.GetRequiredEnv("RDS_PORT")
	name := util.GetRequiredEnv("RDS_DB_NAME")

	log.Println("Connecting to database...")
	db, err := sqlx.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		hostname,
		port,
		username,
		password,
		name,
	))
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Println("Retry database connection in 5 seconds...")
		time.Sleep(time.Duration(5) * time.Second)
		return Open()
	}
	log.Println("Database connected")
	return db, nil
}

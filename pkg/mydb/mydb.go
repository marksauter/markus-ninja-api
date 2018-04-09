package mydb

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/marksauter/markus-ninja-api/pkg/utils"
)

func Open() (*sqlx.DB, error) {
	username := utils.GetRequiredEnv("RDS_USERNAME")
	password := utils.GetRequiredEnv("RDS_PASSWORD")
	hostname := utils.GetRequiredEnv("RDS_HOSTNAME")
	port := utils.GetRequiredEnv("RDS_PORT")
	name := utils.GetRequiredEnv("RDS_DB_NAME")

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

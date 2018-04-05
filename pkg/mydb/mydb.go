package mydb

import (
	"database/sql"
	"log"
	"time"

	"github.com/marksauter/markus-ninja-api/pkg/utils"
)

func Open() (*sql.DB, error) {
	username := utils.GetRequiredEnv("RDS_USERNAME")
	password := utils.GetRequiredEnv("RDS_PASSWORD")
	hostname := utils.GetRequiredEnv("RDS_HOSTNAME")
	port := utils.GetRequiredEnv("RDS_PORT")
	name := utils.GetRequiredEnv("RDS_DB_NAME")

	log.Println("Connecting to database...")
	// Create the database handle, confirm driver is present
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+hostname+":"+port+")/"+name)
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

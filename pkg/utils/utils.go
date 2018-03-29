package utils

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	// "github.com/joho/godotenv"
)

// func LoadEnv() error {
//   // Load env vars from .env
//   return godotenv.Load()
// }

func GetOptionalEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetRequiredEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic("Env variable " + key + " required.")
}

func CheckDatabaseVersion(w http.ResponseWriter, r *http.Request) {
	username := GetRequiredEnv("RDS_USERNAME")
	password := GetRequiredEnv("RDS_PASSWORD")
	hostname := GetRequiredEnv("RDS_HOSTNAME")
	port := GetRequiredEnv("RDS_PORT")
	name := GetRequiredEnv("RDS_DB_NAME")

	// Create the database handle, confirm driver is present
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+hostname+":"+port+")/"+name)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Connect and check the server version
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	switch {
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Fprintf(w, "Connected to: %s", version)
	}
}

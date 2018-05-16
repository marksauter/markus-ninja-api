package myconf

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string

	AWSRegion       string
	AWSUploadBucket string

	DBHost     string
	DBPort     uint16
	DBUser     string
	DBPassword string
	DBName     string

	SMTPHost     string
	SMTPPort     string
	SMTPFromAddr string
	SMTPRootURL  string
	SMTPUser     string
	SMTPPassword string
}

func Load(name string) *Config {
	config := viper.New()
	config.SetConfigName(name)
	config.AddConfigPath("$GOPATH/src/github.com/marksauter/markus-ninja-api/")
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return &Config{
		AppName: config.Get("app.name").(string),

		AWSRegion:       config.Get("aws.region").(string),
		AWSUploadBucket: config.Get("aws.upload_bucket").(string),

		DBHost:     config.Get("db.host").(string),
		DBPort:     uint16(config.Get("db.port").(int64)),
		DBUser:     config.Get("db.user").(string),
		DBPassword: config.Get("db.password").(string),
		DBName:     config.Get("db.name").(string),

		SMTPHost:     config.Get("smtp.host").(string),
		SMTPPort:     config.Get("smtp.port").(string),
		SMTPFromAddr: config.Get("smtp.from_address").(string),
		SMTPRootURL:  config.Get("smtp.root_url").(string),
		SMTPUser:     config.Get("smtp.user").(string),
		SMTPPassword: config.Get("smtp.password").(string),
	}
}

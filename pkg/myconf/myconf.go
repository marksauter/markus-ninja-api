package myconf

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	APIURL    string
	AppName   string
	ClientURL string

	AuthKeyId string

	AWSRegion       string
	AWSUploadBucket string

	DBHost     string
	DBPort     uint16
	DBUser     string
	DBPassword string
	DBName     string

	MailCharSet string
	MailSender  string
	MailRootURL string
}

func Load(name string) *Config {
	config := viper.New()
	config.SetConfigName(name)
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return &Config{
		APIURL:    config.Get("app.api_url").(string),
		AppName:   config.Get("app.name").(string),
		ClientURL: config.Get("app.client_url").(string),

		AuthKeyId: config.Get("auth.key_id").(string),

		AWSRegion:       config.Get("aws.region").(string),
		AWSUploadBucket: config.Get("aws.upload_bucket").(string),

		DBHost:     config.Get("db.host").(string),
		DBPort:     uint16(config.Get("db.port").(int64)),
		DBUser:     config.Get("db.user").(string),
		DBPassword: config.Get("db.password").(string),
		DBName:     config.Get("db.name").(string),

		MailCharSet: config.Get("mail.char_set").(string),
		MailSender:  config.Get("mail.sender").(string),
		MailRootURL: config.Get("mail.root_url").(string),
	}
}

package myconf

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	APIURL    string
	AppName   string
	ClientURL string
	ImagesURL string

	AuthKeyId string

	AWSRegion       string
	AWSUploadBucket string

	DBHost         string
	DBPort         uint16
	DBRootUser     string
	DBRootPassword string
	DBUser         string
	DBPassword     string
	DBName         string

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

	conf := &Config{
		APIURL:    config.Get("app.api_url").(string),
		AppName:   config.Get("app.name").(string),
		ClientURL: config.Get("app.client_url").(string),
		ImagesURL: config.Get("app.images_url").(string),

		AuthKeyId: config.Get("auth.key_id").(string),

		AWSRegion:       config.Get("aws.region").(string),
		AWSUploadBucket: config.Get("aws.upload_bucket").(string),

		DBHost: config.Get("db.host").(string),
		DBPort: uint16(config.Get("db.port").(int64)),
		DBName: config.Get("db.name").(string),

		MailCharSet: config.Get("mail.char_set").(string),
		MailSender:  config.Get("mail.sender").(string),
		MailRootURL: config.Get("mail.root_url").(string),
	}

	dbRootUser := config.Get("db.root_user")
	if dbRootUser != nil {
		conf.DBRootUser = dbRootUser.(string)
	}
	dbRootPassword := config.Get("db.root_password")
	if dbRootPassword != nil {
		conf.DBRootPassword = dbRootPassword.(string)
	}
	dbUser := config.Get("db.user")
	if dbUser != nil {
		conf.DBUser = dbUser.(string)
	}
	dbPassword := config.Get("db.password")
	if dbPassword != nil {
		conf.DBPassword = dbPassword.(string)
	}

	return conf
}

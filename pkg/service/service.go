package service

import (
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
)

type Services struct {
	Auth    *AuthService
	Mail    *MailService
	Storage *StorageService
}

func NewServices(conf *myconf.Config) (*Services, error) {
	authConfig := &AuthServiceConfig{
		KeyID: conf.AuthKeyId,
	}
	mailConfig := &MailServiceConfig{
		CharSet: conf.MailCharSet,
		Sender:  conf.MailSender,
		RootURL: conf.MailRootURL,
	}
	storageSvc, err := NewStorageService()
	if err != nil {
		return nil, err
	}
	return &Services{
		Auth:    NewAuthService(myaws.NewKMS(), authConfig),
		Mail:    NewMailService(myaws.NewSES(), mailConfig),
		Storage: storageSvc,
	}, nil
}

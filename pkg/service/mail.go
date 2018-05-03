package service

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/sirupsen/logrus"
)

type MailServiceConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	FromAddress string
}

func NewMailService(conf *MailServiceConfig) *MailService {
	auth := smtp.PlainAuth("", conf.User, conf.Password, conf.Host)

	return &MailService{
		serverAddr: conf.Host + ":" + conf.Port,
		auth:       auth,
		from:       conf.FromAddress,
	}
}

type MailService struct {
	serverAddr string
	auth       smtp.Auth
	from       string
}

var accountVerificationMailTemplate = template.Must(
	template.New("passwordResetMailTemplate").Parse(
		"To: {{.To}}\r\nSubject: markus ninja account verification\r\n\r\nYour verification code is: {{.Token}}",
	),
)

func (s *MailService) SendAccountVerificationMail(to, token string) error {
	var data = struct {
		To    string
		Token string
	}{
		To:    to,
		Token: token,
	}

	buf := &bytes.Buffer{}
	err := accountVerificationMailTemplate.Execute(buf, data)
	if err != nil {
		return err
	}

	err = smtp.SendMail(s.serverAddr, s.auth, s.from, []string{to}, buf.Bytes())
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"to":    to,
			"error": err,
		}).Error("failed to send account verification email")
	}

	mylog.Log.WithFields(logrus.Fields{
		"to": to,
	}).Info("sent account verification email")

	return nil
}

var passwordResetMailTemplate = template.Must(
	template.New("passwordResetMailTemplate").Parse(
		"To: {{.To}}\r\nSubject: markus ninja password reset\r\n\r\nYour password reset code is: {{.Token}}",
	),
)

func (s *MailService) SendPasswordResetMail(to, token string) error {
	var data = struct {
		To    string
		Token string
	}{
		To:    to,
		Token: token,
	}

	buf := &bytes.Buffer{}
	err := passwordResetMailTemplate.Execute(buf, data)
	if err != nil {
		return err
	}

	err = smtp.SendMail(s.serverAddr, s.auth, s.from, []string{to}, buf.Bytes())
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"to":    to,
			"error": err,
		}).Error("failed to send password reset email")
	}

	mylog.Log.WithFields(logrus.Fields{
		"to": to,
	}).Error("sent password reset email")

	return nil
}

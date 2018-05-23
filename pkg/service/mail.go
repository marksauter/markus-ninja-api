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
	RootURL     string
}

func NewMailService(conf *MailServiceConfig) *MailService {
	auth := smtp.PlainAuth("", conf.User, conf.Password, conf.Host)

	return &MailService{
		rootURL:    conf.RootURL,
		serverAddr: conf.Host + ":" + conf.Port,
		auth:       auth,
		from:       conf.FromAddress,
	}
}

type MailService struct {
	rootURL    string
	serverAddr string
	auth       smtp.Auth
	from       string
}

var emailVerificationMailTemplate = template.Must(
	template.New("passwordResetMailTemplate").Parse(
		"To: {{.To}}\r\n" +
			"Subject: [rkus.ninja] Please verify your email address\r\n\r\n" +
			"Hi @{{.Login}}\r\n\r\n," +
			"Paste the following link into your browser to verify your email address: " +
			"{{.RootURL}}/users/{{.Login}}/emails/confirm_verification/{{.Token}}",
	),
)

func (s *MailService) SendEmailVerificationMail(to, login, emailId, token string) error {
	var data = struct {
		EmailId string
		Login   string
		RootURL string
		To      string
		Token   string
	}{
		EmailId: emailId,
		Login:   login,
		RootURL: s.rootURL,
		To:      to,
		Token:   token,
	}

	buf := &bytes.Buffer{}
	err := emailVerificationMailTemplate.Execute(buf, data)
	if err != nil {
		return err
	}

	err = smtp.SendMail(s.serverAddr, s.auth, s.from, []string{to}, buf.Bytes())
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"to":    to,
			"error": err,
		}).Error("failed to send email verification email")
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"to": to,
	}).Info("sent email verification email")

	return nil
}

var passwordResetMailTemplate = template.Must(
	template.New("passwordResetMailTemplate").Parse(
		"To: {{.To}}\r\n" +
			"Subject: [rkus.ninja] Password reset request\r\n\r\n" +
			"Hi @{{.Login}}\r\n\r\n," +
			"Your password reset code is: {{.Token}}\r\n\r\n" +
			"If you did not request a password reset, then please ignore this message.",
	),
)

func (s *MailService) SendPasswordResetMail(to, login, token string) error {
	var data = struct {
		To    string
		Login string
		Token string
	}{
		To:    to,
		Login: login,
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
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"to": to,
	}).Info("sent password reset email")

	return nil
}

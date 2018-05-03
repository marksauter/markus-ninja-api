package mysmtp

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/sirupsen/logrus"
)

func New(conf *myconf.Config) Mailer {
	auth := smtp.PlainAuth("", conf.SMTPUser, conf.SMTPPassword, conf.SMTPHost)

	mailer := &SMTPMailer{
		ServerAddr: conf.SMTPHost + ":" + conf.SMTPPort,
		Auth:       auth,
		From:       conf.SMTPFromAddr,
		rootURL:    conf.SMTPRootUrl,
	}

	return mailer
}

type Mailer interface {
	SendAccountVerificationMail(to, token string) error
	SendPasswordResetMail(to, token string) error
}

type SMTPMailer struct {
	ServerAddr string
	Auth       smtp.Auth
	From       string
	rootURL    string
}

var accountVerificationMailTemplate = template.Must(
	template.New("passwordResetMailTemplate").Parse(
		"To: {{.To}}\r\nSubject: markus ninja account verification\r\n\r\nClick the following link to verify account: {{.RootURL}}/verify_account?token={{.Token}}",
	),
)

func (m *SMTPMailer) SendAccountVerificationMail(to, token string) error {
	var data = struct {
		RootURL string
		To      string
		Token   string
	}{
		RootURL: m.rootURL,
		To:      to,
		Token:   token,
	}

	buf := &bytes.Buffer{}
	err := accountVerificationMailTemplate.Execute(buf, data)
	if err != nil {
		return err
	}

	err = smtp.SendMail(m.ServerAddr, m.Auth, m.From, []string{to}, buf.Bytes())
	if err != nil {
		mylog.Log.WithFields(logrus.Fields{
			"to":    to,
			"error": err,
		}).Error("failed to send account verification email")
	}

	mylog.Log.WithFields(logrus.Fields{
		"to": to,
	}).Error("sent account verification email")

	return nil
}

var passwordResetMailTemplate = template.Must(
	template.New("passwordResetMailTemplate").Parse(
		"To: {{.To}}\r\nSubject: markus ninja password reset\r\n\r\nClick the following link to reset password: {{.RootURL}}/reset_password?token={{.Token}}",
	),
)

func (m *SMTPMailer) SendPasswordResetMail(to, token string) error {
	var data = struct {
		RootURL string
		To      string
		Token   string
	}{
		RootURL: m.rootURL,
		To:      to,
		Token:   token,
	}

	buf := &bytes.Buffer{}
	err := passwordResetMailTemplate.Execute(buf, data)
	if err != nil {
		return err
	}

	err = smtp.SendMail(m.ServerAddr, m.Auth, m.From, []string{to}, buf.Bytes())
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

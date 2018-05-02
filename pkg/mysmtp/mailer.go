package mysmtp

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
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
	SendPasswordResetMail(to, token string) error
}

type SMTPMailer struct {
	SeverAddr string
	Auth      smtp.Auth
	From      string
	rootURL   string
}

var passwordResetMailTemplate = template.Must(
	template.New("passwordResetMailTemplate").Parse(
		"To: {{.To}}\r\nSubject: markus ninja password reset\r\n\r\nClick the following link to reset password: {{.RootURL}}/#resetPassword?token={{.Token}}",
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
		Token, token,
	}

	buf := &bytes.Buffer{}
	err := passwordResetMailTemplate.Execute(buf, data)
	if err != nil {
		return err
	}

	err = smtp.SendMail(m.ServerAddr, m.Auth, m.From, []string{to}, buf.Bytes())
	if err != nil {
		mylog.Log.WithFields(logurs.Fields{
			"to":    to,
			"error": err,
		}).Error("failed to send password reset email")
	}

	mylog.Log.WithFields(logurs.Fields{
		"to": to,
	}).Error("sent password reset email")
}

package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/sirupsen/logrus"
)

type MailServiceConfig struct {
	CharSet string
	Sender  string
	RootURL string
}

func NewMailService(svc sesiface.SESAPI, conf *MailServiceConfig) *MailService {
	return &MailService{
		conf: conf,
		svc:  svc,
	}
}

type MailService struct {
	conf *MailServiceConfig
	svc  sesiface.SESAPI
}

const (
	EmailVerificationSubject = "[rkus.ninja] Please verify your email address"
	PasswordResetSubject     = "[rkus.ninja] Password reset request"
)

type SendEmailVerificationMailInput struct {
	EmailId   string
	To        string
	Token     string
	UserLogin string
}

func (s *MailService) SendEmailVerificationMail(
	input *SendEmailVerificationMailInput,
) error {
	link := s.conf.RootURL + "/users/" + input.UserLogin + "/emails/" +
		input.EmailId + "/confirm_verification/" + input.Token
	htmlBody := "<p>Hi <strong>@" + input.UserLogin + "</strong>!</p>" +
		"<p>Please verify your email address (" + input.To + "). This will let you start creating.</p>" +
		"<p><a href='" + link + "'>Verify email address</a>.</p>" +
		"<hr>" +
		"<p>Button not working?  Paste the following link into your browser:<br>" +
		"<span>" + link + "</span></p>" +
		"<p>You're receiving this email because you recently created a new " +
		"rkus.ninja account or added a new email address. " +
		"If this wasn't you, please ignore this email.</p>"

	textBody := "Hi @" + input.UserLogin + "!\r\n\r\n" +
		"Please verify your email address (" + input.To + ").  This will let you start creating.\r\n" +
		"Paste the following link into your browser:\r\n" + link +
		"You're receiving this email because you recently created a new " +
		"rkus.ninja account or added a new email address. " +
		"If this wasn't you, please ignore this email."

	sendEmailInput := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(input.To),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(s.conf.CharSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(s.conf.CharSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(s.conf.CharSet),
				Data:    aws.String(EmailVerificationSubject),
			},
		},
		Source: aws.String(s.conf.Sender),
	}

	_, err := s.svc.SendEmail(sendEmailInput)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return err
			}
		}
	}

	mylog.Log.WithFields(logrus.Fields{
		"to": input.To,
	}).Info("sent email verification email")

	return nil
}

type SendPasswordResetInput struct {
	To        string
	Token     string
	UserLogin string
}

func (s *MailService) SendPasswordResetMail(
	input *SendPasswordResetInput,
) error {
	htmlBody := "<p>Hi <strong>@" + input.UserLogin + "</strong>!</p>" +
		"<p>Your password reset code is: " + input.Token + "</p>" +
		"<hr>" +
		"<p>You're receiving this email because you recently requested a " +
		"password reset for your rkus.ninja account. " +
		"If this wasn't you, please ignore this email.</p>"

	textBody := "Hi @" + input.UserLogin + "!\r\n\r\n" +
		"Your password reset code is: " + input.Token + "\r\n\r\n" +
		"You're receiving this email because you recently requested a " +
		"password reset for your rkus.ninja account. " +
		"If this wasn't you, please ignore this email."

	sendEmailInput := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(input.To),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(s.conf.CharSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(s.conf.CharSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(s.conf.CharSet),
				Data:    aws.String(PasswordResetSubject),
			},
		},
		Source: aws.String(s.conf.Sender),
	}

	_, err := s.svc.SendEmail(sendEmailInput)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return err
			}
		}
	}

	mylog.Log.WithFields(logrus.Fields{
		"to": input.To,
	}).Info("sent password reset email")

	return nil
}

package mylog

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
)

var Log = New()

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
	return &Logger{log}
}

func (l *Logger) AccessMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		l.WithFields(logrus.Fields{
			"remote_addr": req.RemoteAddr,
			"method":      req.Method,
			"url":         req.URL,
			"proto":       req.Proto,
		}).Info("Request Info")
		l.WithField("user_agent", req.UserAgent()).Info("")
		if l.Level >= logrus.DebugLevel {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				l.WithField("error", err).Error("Error reading request body")
			}
			reqStr := ioutil.NopCloser(bytes.NewBuffer(body))
			re_escaped := regexp.MustCompile(`\\n|\\`)
			prettyBody := re_escaped.ReplaceAll(body, nil)
			re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
			prettyBody = re_inside_whtsp.ReplaceAll(prettyBody, []byte{' '})
			l.WithField("body", string(prettyBody)).Debug("")
			req.Body = reqStr
		}
		h.ServeHTTP(rw, req)
	})
}

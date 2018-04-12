package mylog

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"

	logging "github.com/op/go-logging"
)

type Logger struct {
	Log       *logging.Logger
	debugMode bool
}

func NewLogger(debugMode bool) *Logger {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	format := logging.MustStringFormatter(
		"%{color}%{time:2006/01/02 15:04:05 -07:00 MST} [%{level:.6s}] %{shortpkg}:%{shortfile}" +
			" : " +
			"%{color:reset}%{message}",
	)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(logging.INFO, "")
	if debugMode {
		backendLeveled.SetLevel(logging.DEBUG, "")
	}

	logging.SetBackend(backendLeveled)
	logger := logging.MustGetLogger("markus-ninja-api")
	return &Logger{Log: logger, debugMode: debugMode}
}

func (l *Logger) AccessMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		l.Log.Infof("%s %s %s %s", req.RemoteAddr, req.Method, req.URL, req.Proto)
		l.Log.Infof("User agent : %s", req.UserAgent())
		if l.debugMode {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				l.Log.Errorf("Reading request body error: %s", err)
			}
			reqStr := ioutil.NopCloser(bytes.NewBuffer(body))
			l.Log.Debugf("Request body : %v", reqStr)
			req.Body = reqStr
		}
		h.ServeHTTP(rw, req)
	})
}

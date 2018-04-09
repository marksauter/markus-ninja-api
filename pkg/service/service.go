package service

import "github.com/marksauter/markus-ninja-api/pkg/mylog"

type Service interface {
	Logger() *mylog.Logger
}

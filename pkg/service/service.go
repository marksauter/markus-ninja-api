package service

import logging "github.com/op/go-logging"

type Service interface {
	log() *logging.Logger
}

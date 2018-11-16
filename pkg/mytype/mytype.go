package mytype

import "errors"

var errUndefined = errors.New("cannot encode status undefined")
var errBadStatus = errors.New("invalid status")

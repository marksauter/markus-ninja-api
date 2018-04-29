package loader

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewPermLoader(svc *data.PermService) *PermLoader {
	return &PermLoader{svc}
}

type PermLoader struct {
	*data.PermService
}

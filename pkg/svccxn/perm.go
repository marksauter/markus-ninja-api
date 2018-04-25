package svccxn

import (
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func NewPermConnection(svc *service.PermService) *PermConnection {
	return &PermConnection{
		svc: svc,
	}
}

type PermConnection struct {
	svc *service.PermService
}

func (r *PermConnection) GetQueryPermission(
	o perm.Operation,
	roles ...string,
) (*perm.QueryPermission, error) {
	return r.svc.GetQueryPermission(o, roles...)
}

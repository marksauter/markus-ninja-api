package myctx

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type ctxService interface {
	NewContext(context.Context, service.Service) (context.Context, bool)
	FromContext(context.Context) (service.Service, bool)
}

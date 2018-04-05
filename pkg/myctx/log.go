package myctx

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/service"
)

var Log = ctxLog{}

type ctxLog struct{}

var logKey key = "log"

func (c *ctxLog) NewContext(ctx context.Context, log *service.Logger) context.Context {
	return context.WithValue(ctx, logKey, log)
}

func (c *ctxLog) FromContext(ctx context.Context) (*service.Logger, bool) {
	log, ok := ctx.Value(logKey).(*service.Logger)
	return log, ok
}
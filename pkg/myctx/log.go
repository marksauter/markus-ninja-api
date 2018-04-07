package myctx

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

var Log = ctxLog{}

type ctxLog struct{}

var logKey key = "log"

func (c *ctxLog) NewContext(ctx context.Context, logger *mylog.Logger) context.Context {
	return context.WithValue(ctx, logKey, logger)
}

func (c *ctxLog) FromContext(ctx context.Context) (*mylog.Logger, bool) {
	log, ok := ctx.Value(logKey).(*mylog.Logger)
	return log, ok
}

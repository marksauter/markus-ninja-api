package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/mygql"
)

type uniformResourceLocatable interface {
	ResourcePath(ctx context.Context) (mygql.URI, error)
	URL(ctx context.Context) (mygql.URI, error)
}

package resolver

import "context"

type studyNode interface {
	Study(ctx context.Context) (*studyResolver, error)
}

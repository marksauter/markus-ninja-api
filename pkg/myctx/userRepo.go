package myctx

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type ctxUserRepo struct {
	ctxRepo
}

var UserRepo = ctxUserRepo{}

var userRepoKey key = "repo.user"

func (cr *ctxUserRepo) NewContext(ctx context.Context, r repo.Repo) (context.Context, bool) {
	userRepo, ok := r.(*repo.UserRepo)
	return context.WithValue(ctx, userRepoKey, userRepo), ok
}

type InvalidFromContextError struct {
	Type reflect.Type
}

func (e *InvalidFromContextError) Error() string {
	if e.Type == nil {
		return "myctx: FromContext(_, nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return fmt.Sprintf("myctx: FromContext(_, non-pointer %s)", e.Type.String())
	}
	return fmt.Sprintf("myctx: FromContext(_, nil %s)", e.Type.String())
}

func (cr *ctxUserRepo) FromContext(ctx context.Context, repo interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	rv := reflect.ValueOf(repo)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidFromContextError{reflect.TypeOf(repo)}
	}
	return nil
}

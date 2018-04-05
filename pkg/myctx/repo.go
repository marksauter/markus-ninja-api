package myctx

import (
	"context"
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type ctxRepo interface {
	NewContext(ctx context.Context, repo repo.Repo) (context.Context, bool)
	FromContext(ctx context.Context) (repo.Repo, bool)
}

func resolveTypeFromId(id string) (string, error) {
	idComponents := strings.Split(id, "_")
	if len(idComponents) != 2 {
		return "", fmt.Errorf(`Invalid id "%v": expected format "Type_id"`, id)
	}
	return idComponents[0], nil
}

func CtxRepoFromId(id string) (ctxRepo, error) {
	t, err := resolveTypeFromId(id)
	if err != nil {
		return nil, err
	}
	switch t {
	case "User":
		return &UserRepo, nil
	default:
		return nil, fmt.Errorf(`No repo found for type "%v"`, t)
	}
}

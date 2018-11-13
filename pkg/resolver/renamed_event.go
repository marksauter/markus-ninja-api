package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type renamedEventResolver struct {
	Conf         *myconf.Config
	Event        *repo.EventPermit
	RenameableID *mytype.OID
	Rename       *data.RenamePayload
	Repos        *repo.Repos
}

func (r *renamedEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *renamedEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *renamedEventResolver) Renameable(ctx context.Context) (*renameableResolver, error) {
	permit, err := r.Repos.GetRenameable(ctx, r.RenameableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	renameable, ok := resolver.(renameable)
	if !ok {
		return nil, errors.New("cannot convert resolver to renameable")
	}
	return &renameableResolver{renameable}, nil
}

func (r *renamedEventResolver) RenamedFrom() string {
	return r.Rename.From
}

func (r *renamedEventResolver) RenamedTo() string {
	return r.Rename.To
}

func (r *renamedEventResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Event.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *renamedEventResolver) User(ctx context.Context) (*userResolver, error) {
	userID, err := r.Event.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

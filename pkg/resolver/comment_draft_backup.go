package resolver

import (
	"context"
	"fmt"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type commentDraftBackupResolver struct {
	Conf               *myconf.Config
	CommentDraftBackup *repo.CommentDraftBackupPermit
	Repos              *repo.Repos
}

func (r *commentDraftBackupResolver) Comment(ctx context.Context) (*commentResolver, error) {
	commentID, err := r.CommentDraftBackup.CommentID()
	if err != nil {
		return nil, err
	}
	comment, err := r.Repos.Comment().Get(ctx, commentID.String)
	if err != nil {
		return nil, err
	}
	return &commentResolver{Comment: comment, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *commentDraftBackupResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.CommentDraftBackup.CreatedAt()
	return graphql.Time{t}, err
}

func (r *commentDraftBackupResolver) Draft() (string, error) {
	return r.CommentDraftBackup.Draft()
}

func (r *commentDraftBackupResolver) ID() (graphql.ID, error) {
	id, err := r.CommentDraftBackup.ID()
	return graphql.ID(fmt.Sprintf("%d", id)), err
}

func (r *commentDraftBackupResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.CommentDraftBackup.UpdatedAt()
	return graphql.Time{t}, err
}

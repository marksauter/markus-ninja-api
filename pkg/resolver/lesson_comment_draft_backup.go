package resolver

import (
	"context"
	"fmt"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type lessonCommentDraftBackupResolver struct {
	Conf                     *myconf.Config
	LessonCommentDraftBackup *repo.LessonCommentDraftBackupPermit
	Repos                    *repo.Repos
}

func (r *lessonCommentDraftBackupResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.LessonCommentDraftBackup.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentDraftBackupResolver) Draft() (string, error) {
	return r.LessonCommentDraftBackup.Draft()
}

func (r *lessonCommentDraftBackupResolver) ID() (graphql.ID, error) {
	id, err := r.LessonCommentDraftBackup.ID()
	return graphql.ID(fmt.Sprintf("%d", id)), err
}

func (r *lessonCommentDraftBackupResolver) LessonComment(ctx context.Context) (*lessonCommentResolver, error) {
	lessonCommentID, err := r.LessonCommentDraftBackup.LessonCommentID()
	if err != nil {
		return nil, err
	}
	lessonComment, err := r.Repos.LessonComment().Get(ctx, lessonCommentID.String)
	if err != nil {
		return nil, err
	}
	return &lessonCommentResolver{LessonComment: lessonComment, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *lessonCommentDraftBackupResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.LessonCommentDraftBackup.UpdatedAt()
	return graphql.Time{t}, err
}

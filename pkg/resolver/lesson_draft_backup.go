package resolver

import (
	"context"
	"fmt"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type lessonDraftBackupResolver struct {
	Conf              *myconf.Config
	LessonDraftBackup *repo.LessonDraftBackupPermit
	Repos             *repo.Repos
}

func (r *lessonDraftBackupResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.LessonDraftBackup.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonDraftBackupResolver) Draft() (string, error) {
	return r.LessonDraftBackup.Draft()
}

func (r *lessonDraftBackupResolver) ID() (graphql.ID, error) {
	id, err := r.LessonDraftBackup.ID()
	return graphql.ID(fmt.Sprintf("%d", id)), err
}

func (r *lessonDraftBackupResolver) Lesson(ctx context.Context) (*lessonResolver, error) {
	lessonID, err := r.LessonDraftBackup.LessonID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().Get(ctx, lessonID.String)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *lessonDraftBackupResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.LessonDraftBackup.UpdatedAt()
	return graphql.Time{t}, err
}

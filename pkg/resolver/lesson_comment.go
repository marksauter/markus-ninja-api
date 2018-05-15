package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type LessonComment = lessonCommentResolver

type lessonCommentResolver struct {
	LessonComment *repo.LessonCommentPermit
	Repos         *repo.Repos
}

func (r *lessonCommentResolver) Author(ctx context.Context) (*userResolver, error) {
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId)
	if err != nil {
		return nil, err
	}
	err = user.ViewerCanAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) Body() (string, error) {
	return r.LessonComment.Body()
}

func (r *lessonCommentResolver) BodyHTML() (mygql.HTML, error) {
	body, err := r.LessonComment.Body()
	if err != nil {
		return "", err
	}
	bodyHTML := util.MarkdownToHTML([]byte(body))
	gqlHTML := mygql.HTML(bodyHTML)
	return gqlHTML, nil
}

func (r *lessonCommentResolver) BodyText() (string, error) {
	body, err := r.LessonComment.Body()
	if err != nil {
		return "", err
	}
	return util.MarkdownToText(body), nil
}

func (r *lessonCommentResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.LessonComment.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) ID() (graphql.ID, error) {
	id, err := r.LessonComment.ID()
	return graphql.ID(id), err
}

func (r *lessonCommentResolver) Lesson() (*lessonResolver, error) {
	lessonId, err := r.LessonComment.LessonId()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().Get(lessonId)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) PublishedAt() (*graphql.Time, error) {
	t, err := r.LessonComment.PublishedAt()
	if err != nil {
		return nil, err
	}
	return &graphql.Time{t}, nil
}

func (r *lessonCommentResolver) ResourcePath(ctx context.Context) (mygql.URI, error) {
	var uri mygql.URI
	study, err := r.Study()
	if err != nil {
		return uri, err
	}
	studyResourcePath, err := study.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	lessonId, err := r.LessonComment.ID()
	if err != nil {
		return uri, err
	}
	_, err = r.Repos.Lesson().AddPermission(perm.ReadLesson)
	if err != nil {
		return uri, err
	}
	lesson, err := r.Repos.Lesson().Get(lessonId)
	if err != nil {
		return uri, err
	}
	lessonNumber, err := lesson.Number()
	if err != nil {
		return uri, err
	}
	createdAt, err := r.CreatedAt()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/lesson/%s#%s", studyResourcePath, lessonNumber, createdAt.Unix()))
	return uri, nil
}

func (r *lessonCommentResolver) Study() (*studyResolver, error) {
	studyId, err := r.LessonComment.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(studyId)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.LessonComment.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) URL(ctx context.Context) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *lessonCommentResolver) ViewerCanDelete(ctx context.Context) (bool, error) {
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

func (r *lessonCommentResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

func (r *lessonCommentResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

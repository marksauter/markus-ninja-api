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

var clientURL = "http://localhost:3000"

type Lesson = lessonResolver

type lessonResolver struct {
	Lesson *repo.LessonPermit
	Repos  *repo.Repos
}

func (r *lessonResolver) Author(ctx context.Context) (*userResolver, error) {
	userId, err := r.Lesson.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId.String)
	if err != nil {
		return nil, err
	}
	err = user.ViewerCanAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *lessonResolver) Body() (string, error) {
	return r.Lesson.Body()
}

func (r *lessonResolver) BodyHTML() (mygql.HTML, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	bodyHTML := util.MarkdownToHTML([]byte(body))
	gqlHTML := mygql.HTML(bodyHTML)
	return gqlHTML, nil
}

func (r *lessonResolver) BodyText() (string, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	return util.MarkdownToText(body), nil
}

func (r *lessonResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Lesson.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) ID() (graphql.ID, error) {
	id, err := r.Lesson.ID()
	return graphql.ID(id.String), err
}

func (r *lessonResolver) Number() (int32, error) {
	return r.Lesson.Number()
}

func (r *lessonResolver) PublishedAt() (graphql.Time, error) {
	t, err := r.Lesson.PublishedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) ResourcePath(ctx context.Context) (mygql.URI, error) {
	var uri mygql.URI
	study, err := r.Study(ctx)
	if err != nil {
		return uri, err
	}
	studyResourcePath, err := study.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	number, err := r.Number()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/lesson/%d", studyResourcePath, number))
	return uri, nil
}

func (r *lessonResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyId, err := r.Lesson.StudyId()
	if err != nil {
		return nil, err
	}
	_, err = r.Repos.Study().AddPermission(perm.ReadStudy)
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *lessonResolver) Title() (string, error) {
	return r.Lesson.Title()
}

func (r *lessonResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Lesson.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) URL(ctx context.Context) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *lessonResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.Lesson.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

func (r *lessonResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.Lesson.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

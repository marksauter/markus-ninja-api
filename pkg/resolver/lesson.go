package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

var clientURL = "http://localhost:3000"

type Lesson = lessonResolver

type lessonResolver struct {
	lesson *repo.LessonPermit
	Repos  *repo.Repos
}

func (r *lessonResolver) Author() (*userResolver, error) {
	userId, err := r.lesson.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId)
	if err != nil {
		return nil, err
	}
	return &userResolver{user}, nil
}

func (r *lessonResolver) Body() (string, error) {
	return r.Body()
}

func (r *lessonResolver) BioHTML() (mygql.HTML, error) {
	body, err := r.Body()
	if err != nil {
		return "", err
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", body))
	return h, nil
}

func (r *lessonResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.lesson.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) ID() (graphql.ID, error) {
	id, err := r.lesson.ID()
	return graphql.ID(id), err
}

func (r *lessonResolver) LastEditedAt() (graphql.Time, error) {
	t, err := r.lesson.LastEditedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) Number() (int32, error) {
	return r.lesson.Number()
}

func (r *lessonResolver) PublishedAt() (graphql.Time, error) {
	t, err := r.lesson.PublishedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) Study() (*studyResolver, error) {
	studyId, err := r.lesson.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(studyId)
	if err != nil {
		return nil, err
	}
	return &studyResolver{study}, nil
}

func (r *lessonResolver) Title() (string, error) {
	return r.lesson.Title()
}

func (r *lessonResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	login, err := r.lesson.Login()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("/%s", login))
	return uri, nil
}

func (r *lessonResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	login, err := r.lesson.Login()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, login))
	return uri, nil
}

func (r *lessonResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	id, err := r.user.ID()
	if err != nil {
		return false, err
	}
	viewerId, _ := viewer.ID()
	userId, err := r.lesson.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId
}

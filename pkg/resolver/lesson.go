package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type lessonResolver struct {
	Lesson *repo.LessonPermit
	Repos  *repo.Repos
}

func (r *lessonResolver) Author() (*userResolver, error) {
	userId, err := r.Lesson.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *lessonResolver) Body() (string, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	return body.String, nil
}

func (r *lessonResolver) BodyHTML() (mygql.HTML, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	return mygql.HTML(body.ToHTML()), nil
}

func (r *lessonResolver) BodyText() (string, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	return body.ToText(), nil
}

func (r *lessonResolver) Comments(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*lessonCommentConnectionResolver, error) {
	userId, err := r.Lesson.UserId()
	if err != nil {
		return nil, err
	}
	studyId, err := r.Lesson.StudyId()
	if err != nil {
		return nil, err
	}
	lessonId, err := r.Lesson.ID()
	if err != nil {
		return nil, err
	}
	lessonCommentOrder, err := ParseLessonCommentOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		lessonCommentOrder,
	)
	if err != nil {
		return nil, err
	}

	lessonComments, err := r.Repos.LessonComment().GetByLesson(
		userId.String,
		studyId.String,
		lessonId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.LessonComment().CountByLesson(
		userId.String,
		studyId.String,
		lessonId.String,
	)
	if err != nil {
		return nil, err
	}
	lessonCommentConnectionResolver, err := NewLessonCommentConnectionResolver(
		lessonComments,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return lessonCommentConnectionResolver, nil
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

func (r *lessonResolver) Events(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*eventConnectionResolver, error) {
	lessonId, err := r.Lesson.ID()
	if err != nil {
		return nil, err
	}
	eventOrder, err := ParseEventOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		eventOrder,
	)
	if err != nil {
		return nil, err
	}

	events, err := r.Repos.Event().GetByTarget(
		lessonId.String,
		pageOptions,
		data.OnlyMentionEvents,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Event().CountByTarget(
		lessonId.String,
		data.OnlyMentionEvents,
	)
	if err != nil {
		return nil, err
	}
	eventConnectionResolver, err := NewEventConnectionResolver(
		events,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return eventConnectionResolver, nil
}

func (r *lessonResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	userLogin, err := r.Lesson.UserLogin()
	if err != nil {
		return uri, err
	}
	studyName, err := r.Lesson.StudyName()
	if err != nil {
		return uri, err
	}
	number, err := r.Number()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s/lesson/%d", userLogin, studyName, number))
	return uri, nil
}

func (r *lessonResolver) Study() (*studyResolver, error) {
	studyId, err := r.Lesson.StudyId()
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

func (r *lessonResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, resourcePath))
	return uri, nil
}

func (r *lessonResolver) ViewerCanUpdate() bool {
	lesson := r.Lesson.Get()
	return r.Repos.Lesson().ViewerCanUpdate(lesson)
}

func (r *lessonResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userId, err := r.Lesson.UserId()
	if err != nil {
		return false, err
	}

	return viewer.Id.String == userId.String, nil
}

func (r *lessonResolver) ViewerHasEnrolled(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	lessonId, err := r.Lesson.ID()
	if err != nil {
		return false, err
	}

	if _, err := r.Repos.LessonEnroll().Get(lessonId.String, viewer.Id.String); err != nil {
		if err == data.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

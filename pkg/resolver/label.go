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

type Label = labelResolver

type labelResolver struct {
	Label *repo.LabelPermit
	Repos *repo.Repos
}

func (r *labelResolver) Color() (string, error) {
	return r.Label.Color()
}

func (r *labelResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Label.CreatedAt()
	return graphql.Time{t}, err
}

func (r *labelResolver) Description() (string, error) {
	return r.Label.Description()
}

func (r *labelResolver) ID() (graphql.ID, error) {
	id, err := r.Label.ID()
	return graphql.ID(id.String), err
}

func (r *labelResolver) IsDefault() (bool, error) {
	return r.Label.IsDefault()
}

func (r *labelResolver) Lessons(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*lessonConnectionResolver, error) {
	lessonOrder, err := ParseLessonOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		lessonOrder,
	)
	if err != nil {
		return nil, err
	}

	labelId, err := r.Label.ID()
	if err != nil {
		return nil, err
	}
	users, err := r.Repos.Lesson().GetByLabel(
		labelId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Lesson().CountByLabel(
		labelId.String,
	)
	if err != nil {
		return nil, err
	}
	lessonConnectionResolver, err := NewLessonConnectionResolver(
		users,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return lessonConnectionResolver, nil
}

func (r *labelResolver) Name() (string, error) {
	return r.Label.Name()
}

func (r *labelResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	name, err := r.Name()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("labels/%s", name))
	return uri, nil
}

func (r *labelResolver) Study() (*studyResolver, error) {
	studyId, err := r.Label.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *labelResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Label.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *labelResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, resourcePath))
	return uri, nil
}

func (r *labelResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	for _, role := range viewer.Roles.Elements {
		if role.String == data.AdminRole {
			return true, nil
		}
	}
	return false, nil
}

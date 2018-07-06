package resolver

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
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
	idStr := ""
	if id != nil {
		idStr = id.String
	}
	return graphql.ID(idStr), err
}

func (r *labelResolver) IsDefault() (bool, error) {
	return r.Label.IsDefault()
}

func (r *labelResolver) Labelables(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
		Type    string
	},
) (*labelableConnectionResolver, error) {
	labelableType, err := ParseLabelableType(args.Type)
	if err != nil {
		return nil, err
	}
	labelableOrder, err := ParseLabelableOrder(labelableType, args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		labelableOrder,
	)
	if err != nil {
		return nil, err
	}

	id, err := r.Label.ID()
	if err != nil {
		return nil, err
	}

	lessonCount, err := r.Repos.Lesson().CountByLabel(id.String)
	if err != nil {
		return nil, err
	}
	permits := []repo.Permit{}

	switch labelableType {
	case LabelableTypeLesson:
		studies, err := r.Repos.Lesson().GetByLabel(id.String, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.Permit, len(studies))
		for i, l := range studies {
			permits[i] = l
		}
	default:
		return nil, fmt.Errorf("invalid type %s for labelable type", labelableType.String())
	}

	return NewLabelableConnectionResolver(
		r.Repos,
		permits,
		pageOptions,
		lessonCount,
	)
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

func (r *labelResolver) ViewerCanDelete() bool {
	label := r.Label.Get()
	return r.Repos.Label().ViewerCanDelete(label)
}

func (r *labelResolver) ViewerCanUpdate() bool {
	label := r.Label.Get()
	return r.Repos.Label().ViewerCanUpdate(label)
}

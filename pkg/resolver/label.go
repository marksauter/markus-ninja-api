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
		Search  *string
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

	permits := []repo.NodePermit{}

	switch labelableType {
	case LabelableTypeLesson:
		filters := &data.LessonFilterOptions{
			Search: args.Search,
		}
		lessons, err := r.Repos.Lesson().GetByLabel(ctx, id.String, pageOptions, filters)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(lessons))
		for i, l := range lessons {
			permits[i] = l
		}
	case LabelableTypeLessonComment:
		lessonComments, err := r.Repos.LessonComment().GetByLabel(ctx, id.String, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(lessonComments))
		for i, l := range lessonComments {
			permits[i] = l
		}
	default:
		return nil, fmt.Errorf("invalid type %s for labelable type", labelableType.String())
	}

	return NewLabelableConnectionResolver(
		r.Repos,
		permits,
		pageOptions,
		id,
		args.Search,
	)
}

func (r *labelResolver) Name() (string, error) {
	return r.Label.Name()
}

func (r *labelResolver) ResourcePath(ctx context.Context) (mygql.URI, error) {
	var uri mygql.URI
	study, err := r.Study(ctx)
	if err != nil {
		return uri, err
	}
	studyPath, err := study.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	name, err := r.Name()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/labels/%s", string(studyPath), name))
	return uri, nil
}

func (r *labelResolver) Study(
	ctx context.Context,
) (*studyResolver, error) {
	studyID, err := r.Label.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *labelResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Label.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *labelResolver) URL(ctx context.Context) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *labelResolver) ViewerCanDelete(
	ctx context.Context,
) bool {
	label := r.Label.Get()
	return r.Repos.Label().ViewerCanDelete(ctx, label)
}

func (r *labelResolver) ViewerCanUpdate(
	ctx context.Context,
) bool {
	label := r.Label.Get()
	return r.Repos.Label().ViewerCanUpdate(ctx, label)
}

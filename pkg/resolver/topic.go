package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type topicResolver struct {
	Conf  *myconf.Config
	Repos *repo.Repos
	Topic *repo.TopicPermit
}

func (r *topicResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Topic.CreatedAt()
	return graphql.Time{t}, err
}

func (r *topicResolver) Description() (string, error) {
	return r.Topic.Description()
}

func (r *topicResolver) ID() (graphql.ID, error) {
	id, err := r.Topic.ID()
	return graphql.ID(id.String), err
}

func (r *topicResolver) Name() (string, error) {
	return r.Topic.Name()
}

func (r *topicResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	name, err := r.Name()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("/topics/%s", name))
	return uri, nil
}

func (r *topicResolver) Topicables(
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
) (*topicableConnectionResolver, error) {
	topicableType, err := ParseTopicableType(args.Type)
	if err != nil {
		return nil, err
	}
	topicableOrder, err := ParseTopicableOrder(topicableType, args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		topicableOrder,
	)
	if err != nil {
		return nil, err
	}

	id, err := r.Topic.ID()
	if err != nil {
		return nil, err
	}

	permits := []repo.NodePermit{}

	switch topicableType {
	case TopicableTypeCourse:
		filters := &data.CourseFilterOptions{
			Search: args.Search,
		}
		courses, err := r.Repos.Course().GetByTopic(ctx, id.String, pageOptions, filters)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(courses))
		for i, l := range courses {
			permits[i] = l
		}
	case TopicableTypeStudy:
		filters := &data.StudyFilterOptions{
			Search: args.Search,
		}
		studies, err := r.Repos.Study().GetByTopic(ctx, id.String, pageOptions, filters)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(studies))
		for i, l := range studies {
			permits[i] = l
		}
	default:
		return nil, fmt.Errorf("invalid type %s for topicable type", topicableType.String())
	}

	return NewTopicableConnectionResolver(
		permits,
		pageOptions,
		id,
		args.Search,
		r.Repos,
		r.Conf,
	)
}

func (r *topicResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Topic.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *topicResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", r.Conf.ClientURL, resourcePath))
	return uri, nil
}

func (r *topicResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
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

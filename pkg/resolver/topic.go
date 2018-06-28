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

type Topic = topicResolver

type topicResolver struct {
	Topic *repo.TopicPermit
	Repos *repo.Repos
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
	uri = mygql.URI(fmt.Sprintf("topics/%s", name))
	return uri, nil
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
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, resourcePath))
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

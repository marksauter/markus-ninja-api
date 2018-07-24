package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type referencedEventResolver struct {
	Event *repo.EventPermit
	Repos *repo.Repos
}

func (r *referencedEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *referencedEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *referencedEventResolver) IsCrossStudy(
	ctx context.Context,
) (bool, error) {
	sourceId, err := r.Event.SourceId()
	if err != nil {
		return false, err
	}
	sourcePermit, err := r.Repos.GetEventSourceable(ctx, sourceId)
	if err != nil {
		return false, err
	}
	source, ok := sourcePermit.(repo.StudyNodePermit)
	if !ok {
		return false, errors.New("cannot convert source_permit to study_node_permit")
	}
	sourceStudyId, err := source.StudyId()
	if err != nil {
		return false, err
	}
	targetId, err := r.Event.TargetId()
	if err != nil {
		return false, err
	}
	targetPermit, err := r.Repos.GetEventTargetable(ctx, targetId)
	if err != nil {
		return false, err
	}
	target, ok := targetPermit.(repo.StudyNodePermit)
	if !ok {
		return false, errors.New("cannot convert target_permit to study_node_permit")
	}
	targetStudyId, err := target.StudyId()
	if err != nil {
		return false, err
	}
	return sourceStudyId.String != targetStudyId.String, nil
}

func (r *referencedEventResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	id, err := r.Event.SourceId()
	if err != nil {
		return uri, err
	}
	permit, err := r.Repos.GetEventSourceable(ctx, id)
	if err != nil {
		return uri, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return uri, err
	}
	urlable, ok := resolver.(uniformResourceLocatable)
	if !ok {
		return uri, errors.New("cannot convert resolver to uniform_resource_locatable")
	}
	return urlable.ResourcePath(ctx)
}

func (r *referencedEventResolver) Source(
	ctx context.Context,
) (*eventSourceableResolver, error) {
	id, err := r.Event.SourceId()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetEventSourceable(ctx, id)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	eventSourceable, ok := resolver.(eventSourceable)
	if !ok {
		return nil, errors.New("cannot convert resolver to event sourceable")
	}
	return &eventSourceableResolver{eventSourceable}, nil
}

func (r *referencedEventResolver) Target(ctx context.Context) (*eventTargetableResolver, error) {
	id, err := r.Event.TargetId()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetEventTargetable(ctx, id)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	eventTargetable, ok := resolver.(eventTargetable)
	if !ok {
		return nil, errors.New("cannot convert resolver to event targetable")
	}
	return &eventTargetableResolver{eventTargetable}, nil
}

func (r *referencedEventResolver) URL(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *referencedEventResolver) User(ctx context.Context) (*userResolver, error) {
	userId, err := r.Event.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

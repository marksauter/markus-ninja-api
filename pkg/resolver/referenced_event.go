package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type referencedEventResolver struct {
	Event          *repo.EventPermit
	ReferenceableId *mytype.OID
	Repos          *repo.Repos
	SourceId       *mytype.OID
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
	studyId, err := r.Event.StudyId()
	if err != nil {
		return false, err
	}
	source, err := r.Repos.Lesson().Get(ctx, r.SourceId.String)
	if err != nil {
		return false, err
	}
	sourceStudyId, err := source.StudyId()
	if err != nil {
		return false, err
	}
	return studyId.String != sourceStudyId.String, nil
}

func (r *referencedEventResolver) Referenceable(ctx context.Context) (*referenceableResolver, error) {
	permit, err := r.Repos.GetReferenceable(ctx, r.ReferenceableId)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	referenceable, ok := resolver.(referenceable)
	if !ok {
		return nil, errors.New("cannot convert resolver to referenceable")
	}
	return &referenceableResolver{referenceable}, nil
}

func (r *referencedEventResolver) Source(ctx context.Context) (*lessonResolver, error) {
	source, err := r.Repos.Lesson().Get(ctx, r.SourceId.String)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: source, Repos: r.Repos}, nil
}

func (r *referencedEventResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyId, err := r.Event.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
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

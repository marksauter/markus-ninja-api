package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type referencedEventResolver struct {
	Conf            *myconf.Config
	Event           *repo.EventPermit
	ReferenceableID *mytype.OID
	Repos           *repo.Repos
	SourceID        *mytype.OID
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
	studyID, err := r.Event.StudyID()
	if err != nil {
		return false, err
	}
	source, err := r.Repos.Lesson().Get(ctx, r.SourceID.String)
	if err != nil {
		return false, err
	}
	sourceStudyID, err := source.StudyID()
	if err != nil {
		return false, err
	}
	return studyID.String != sourceStudyID.String, nil
}

func (r *referencedEventResolver) Referenceable(ctx context.Context) (*referenceableResolver, error) {
	permit, err := r.Repos.GetReferenceable(ctx, r.ReferenceableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
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
	source, err := r.Repos.Lesson().Get(ctx, r.SourceID.String)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: source, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *referencedEventResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Event.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *referencedEventResolver) User(ctx context.Context) (*userResolver, error) {
	userID, err := r.Event.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

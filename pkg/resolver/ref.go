package resolver

import (
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Ref = refResolver

type refResolver struct {
	Ref   *repo.RefPermit
	Repos *repo.Repos
}

func (r *refResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Ref.CreatedAt()
	return graphql.Time{t}, err
}

func (r *refResolver) ID() (graphql.ID, error) {
	id, err := r.Ref.ID()
	return graphql.ID(id.String), err
}

func (r *refResolver) Target() (*referenceTargetResolver, error) {
	id, err := r.Ref.TargetId()
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Lesson":
		lesson, err := r.Repos.Lesson().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &referenceTargetResolver{Subject: lesson, Repos: r.Repos}, nil
	case "User":
		user, err := r.Repos.User().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &referenceTargetResolver{Subject: user, Repos: r.Repos}, nil
	default:
		return nil, errors.New("invalid target id")
	}
}

func (r *refResolver) Source() (*referenceSourceResolver, error) {
	id, err := r.Ref.SourceId()
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Lesson":
		lesson, err := r.Repos.Lesson().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &referenceSourceResolver{Subject: lesson, Repos: r.Repos}, nil
	case "LessonComment":
		lessonComment, err := r.Repos.LessonComment().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &referenceSourceResolver{Subject: lessonComment, Repos: r.Repos}, nil
	default:
		return nil, errors.New("invalid source id")
	}
}

func (r *refResolver) User() (*userResolver, error) {
	userId, err := r.Ref.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

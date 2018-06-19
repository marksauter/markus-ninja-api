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

func (r *refResolver) Referent() (*nodeResolver, error) {
	id, err := r.Ref.ReferentId()
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Lesson":
		lesson, err := r.Repos.Lesson().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&lessonResolver{Lesson: lesson, Repos: r.Repos}}, nil
	case "User":
		user, err := r.Repos.User().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&userResolver{User: user, Repos: r.Repos}}, nil
	default:
		return nil, errors.New("invalid referent id")
	}
}

func (r *refResolver) Referrer() (*nodeResolver, error) {
	id, err := r.Ref.ReferrerId()
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Lesson":
		lesson, err := r.Repos.Lesson().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&lessonResolver{Lesson: lesson, Repos: r.Repos}}, nil
	case "LessonComment":
		lessonComment, err := r.Repos.LessonComment().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&lessonCommentResolver{LessonComment: lessonComment, Repos: r.Repos}}, nil
	default:
		return nil, errors.New("invalid referrer id")
	}
}

func (r *refResolver) Study() (*studyResolver, error) {
	studyId, err := r.Ref.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

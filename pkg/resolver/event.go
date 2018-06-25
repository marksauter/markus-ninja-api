package resolver

import (
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Event = eventResolver

type eventResolver struct {
	Event *repo.EventPermit
	Repos *repo.Repos
}

func (r *eventResolver) Action() (string, error) {
	return r.Event.Action()
}

func (r *eventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *eventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *eventResolver) Source() (*eventSourceResolver, error) {
	id, err := r.Event.SourceId()
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Lesson":
		lesson, err := r.Repos.Lesson().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &eventSourceResolver{Subject: lesson, Repos: r.Repos}, nil
	case "LessonComment":
		lessonComment, err := r.Repos.LessonComment().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &eventSourceResolver{Subject: lessonComment, Repos: r.Repos}, nil
	case "Study":
		study, err := r.Repos.Study().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &eventSourceResolver{Subject: study, Repos: r.Repos}, nil
	case "User":
		user, err := r.Repos.User().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &eventSourceResolver{Subject: user, Repos: r.Repos}, nil
	default:
		return nil, errors.New("invalid source id")
	}
}

func (r *eventResolver) Target() (*eventTargetResolver, error) {
	id, err := r.Event.TargetId()
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Lesson":
		lesson, err := r.Repos.Lesson().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &eventTargetResolver{Subject: lesson, Repos: r.Repos}, nil
	case "Study":
		study, err := r.Repos.Study().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &eventTargetResolver{Subject: study, Repos: r.Repos}, nil
	case "User":
		user, err := r.Repos.User().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &eventTargetResolver{Subject: user, Repos: r.Repos}, nil
	default:
		return nil, errors.New("invalid target id")
	}
}

func (r *eventResolver) User() (*userResolver, error) {
	userId, err := r.Event.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

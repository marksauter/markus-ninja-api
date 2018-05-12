package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Study = studyResolver

type studyResolver struct {
	Study *repo.StudyPermit
	Repos *repo.Repos
}

func (r *studyResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Study.CreatedAt()
	return graphql.Time{t}, err
}

func (r *studyResolver) Description() (string, error) {
	return r.Description()
}

func (r *studyResolver) DescriptionHTML() (mygql.HTML, error) {
	description, err := r.Description()
	if err != nil {
		return "", err
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", description))
	return h, nil
}

func (r *studyResolver) ID() (graphql.ID, error) {
	id, err := r.Study.ID()
	return graphql.ID(id), err
}

func (r *studyResolver) Lesson(
	ctx context.Context,
	args struct{ Number int32 },
) (*lessonResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.ReadLesson,
		viewer.Roles()...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error retrieving query permission")
		return nil, repo.ErrAccessDenied
	}
	r.Repos.Lesson().AddPermission(queryPerm)
	id, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().GetByStudyNumber(id, args.Number)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}

type PageDirection int

const (
	ForwardPagination PageDirection = iota
	BackwardPagination
)

type Cursor struct {
	s *string
	v *string
}

func NewCursor(cursor *string) (*Cursor, error) {
	v, err := data.DecodeCursor(cursor)
	if err != nil {
		return nil, err
	}
	c := &Cursor{
		s: cursor,
		v: v,
	}
	return c, nil
}

func (c *Cursor) String() string {
	if c.s != nil {
		return *c.s
	}
	return ""
}

func (c *Cursor) Value() string {
	if c.v != nil {
		return *c.v
	}
	return ""
}

type Pagination struct {
	After     *Cursor
	Before    *Cursor
	Direction PageDirection
	first     int32
	last      int32
}

func NewPagination(after, before *string, first, last *int32) (*Pagination, error) {
	pagination := &Pagination{}
	if first == nil && last == nil {
		return nil, fmt.Errorf("You must provide a `first` or `last` value to properly paginate.")
	} else if first != nil {
		if last != nil {
			return nil, fmt.Errorf("Passing both `first` and `last` values to paginate the connection is not supported.")
		}
		pagination.first = *first
		pagination.Direction = ForwardPagination
	} else {
		pagination.last = *last
		pagination.Direction = BackwardPagination
	}
	if after != nil {
		a, err := NewCursor(after)
		if err != nil {
			return nil, err
		}
		pagination.After = a
	}
	if before != nil {
		b, err := NewCursor(before)
		if err != nil {
			return nil, err
		}
		pagination.Before = b
	}
	return pagination, nil
}

func (p *Pagination) Limit() int32 {
	// Assuming one of these is 0, so the sum will be the non-zero field
	return p.first + p.last
}

func (r *studyResolver) Lessons(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *LessonOrderArg
	},
) (*lessonConnectionResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.ReadLesson,
		viewer.Roles()...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error retrieving query permission")
		return nil, repo.ErrAccessDenied
	}
	r.Repos.Lesson().AddPermission(queryPerm)
	id, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lessonOrder, err := ParseLessonOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pagination, err := NewPagination(args.After, args.Before, args.First, args.Last)
	if err != nil {
		return nil, err
	}

	pageOptions := &data.PageOptions{
		After:  pagination.After.Value(),
		Before: pagination.Before.Value(),
		Order:  lessonOrder,
		Limit:  pagination.Limit(),
	}

	lessons, err := r.Repos.Lesson().GetByStudyId(id, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Lesson().CountByStudy(id)
	if err != nil {
		return nil, err
	}
	lessonConnectionResolver, err := NewLessonConnectionResolver(
		lessons,
		pagination,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return lessonConnectionResolver, nil
}

func (r *studyResolver) LessonCount() (int32, error) {
	id, err := r.Study.ID()
	if err != nil {
		var count int32
		return count, err
	}
	return r.Repos.Lesson().CountByStudy(id)
}

func (r *studyResolver) Name() (string, error) {
	return r.Study.Name()
}

func (r *studyResolver) NameWithOwner() (string, error) {
	name, err := r.Name()
	if err != nil {
		return "", err
	}
	owner, err := r.Owner()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", owner, name), nil
}

func (r *studyResolver) Owner() (*userResolver, error) {
	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *studyResolver) PublishedAt() (graphql.Time, error) {
	t, err := r.Study.PublishedAt()
	return graphql.Time{t}, err
}

func (r *studyResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	nameWithOwner, err := r.NameWithOwner()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("/%s", nameWithOwner))
	return uri, nil
}

func (r *studyResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Study.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *studyResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *studyResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.Study.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

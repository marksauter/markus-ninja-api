package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
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

func (r *studyResolver) Lessons(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *LessonOrderArg
	},
) ([]*lessonResolver, error) {
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
	pageOptions := &data.PageOptions{
		Direction: lessonOrder.Direction,
		Field:     lessonOrder.Field,
		Limit:     ,
		Relation:  data.GreaterThan,
	}
	if args.First == nil && args.Last == nil {
		return nil, fmt.Errorf("You must provide a first or last value to properly paginate the `lessons`")
	} else if args.First != nil {
		pageOptions.Limit = args.First 
	} else if args.Last != nil {
		pageOptions.Limit = args.Last 
	}
	if args.After != nil {
		pageOptions.Value().Set(args.After)
	}
	lessons, err := r.Repos.Lesson().GetByStudyId(id, pageOptions)
	if err != nil {
		return nil, err
	}
	lessonResolvers := make([]*lessonResolver, len(lessons))
	for i, l := range lessons {
		lessonResolvers[i] = &lessonResolver{Lesson: l, Repos: r.Repos}
	}
	return lessonResolvers, nil
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

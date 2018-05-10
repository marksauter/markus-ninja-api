package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

var clientURL = "http://localhost:3000"

type Lesson = lessonResolver

type lessonResolver struct {
	Lesson *repo.LessonPermit
	Repos  *repo.Repos
}

func (r *lessonResolver) Author() (*userResolver, error) {
	userId, err := r.Lesson.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *lessonResolver) Body() (string, error) {
	return r.Body()
}

func (r *lessonResolver) BodyHTML() (mygql.HTML, error) {
	body, err := r.Body()
	if err != nil {
		return "", err
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", body))
	return h, nil
}

func (r *lessonResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Lesson.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) ID() (graphql.ID, error) {
	id, err := r.Lesson.ID()
	return graphql.ID(id), err
}

func (r *lessonResolver) LastEditedAt() (graphql.Time, error) {
	t, err := r.Lesson.LastEditedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) Number() (int32, error) {
	return r.Lesson.Number()
}

func (r *lessonResolver) PublishedAt() (graphql.Time, error) {
	t, err := r.Lesson.PublishedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	study, err := r.Study()
	if err != nil {
		return uri, err
	}
	studyResourcePath, err := study.ResourcePath()
	if err != nil {
		return uri, err
	}
	number, err := r.Number()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/lesson/%s", studyResourcePath, number))
	return uri, nil
}

func (r *lessonResolver) Study() (*studyResolver, error) {
	studyId, err := r.Lesson.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(studyId)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *lessonResolver) Title() (string, error) {
	return r.Lesson.Title()
}

func (r *lessonResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *lessonResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.Lesson.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

func (r *lessonResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.Lesson.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

type lessonEdgeResolver struct {
	cursor string
	node   *repo.LessonPermit
	repos  *repo.Repos
}

func (r *lessonEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *lessonEdgeResolver) Node() *nodeResolver {
	return &nodeResolver{&lessonResolver{Lesson: r.node, Repos: r.repos}}
}

type lessonConnectionResolver struct {
	lessons    []*repo.LessonPermit
	totalCount int32
	repos      *repo.Repos
}

func (r *lessonConnectionResolver) Edges() *[]*lessonEdgeResolver {
	edges := make([]*lessonEdgeResolver, len(r.lessons))
	for i := range edges {
		lid, _ := r.lessons[i].ID()
		edges[i] = &lessonEdgeResolver{
			cursor: lid,
			node:   r.lessons[i],
		}
	}
	return &edges
}

func (r *lessonConnectionResolver) Nodes() *[]*lessonResolver {
	nodes := make([]*lessonResolver, len(r.lessons))
	for i := range nodes {
		nodes[i] = &lessonResolver{Lesson: r.lessons[i], Repos: r.repos}
	}
	return &nodes
}

func (r *lessonConnectionResolver) PageInfo() *pageInfoResolver {
	return &pageInfoResolver{}
}

func (r *lessonConnectionResolver) TotalCount() int32 {
	return r.totalCount
}

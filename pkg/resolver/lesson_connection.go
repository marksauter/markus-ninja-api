package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonConnectionResolver(
	cursor *string,
	lessons []*repo.LessonPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
) *lessonConnectionResolver {
	return &lessonConnectionResolver{
		cursor:      cursor,
		lessons:     lessons,
		pageOptions: pageOptions,
		totalCount:  totalCount,
		repos:       repos,
	}
}

type lessonConnectionResolver struct {
	cursor      *string
	lessons     []*repo.LessonPermit
	pageOptions *data.PageOptions
	totalCount  int32
	repos       *repo.Repos
}

func (r *lessonConnectionResolver) Edges() (*[]*lessonEdgeResolver, error) {
	edges := make([]*lessonEdgeResolver, len(r.lessons))
	fieldName := r.pageOptions.Field.Name()
	for i := range edges {
		cursor, err := r.pageOptions.Field.EncodeCursor(r.lessons[i])
		if err != nil {
			return nil, err
		}
		lessonEdge := NewLessonEdgeResolver(cursor, r.lessons[i], r.repos)
		edges[i] = lessonEdge
	}
	return &edges, nil
}

func (r *lessonConnectionResolver) Nodes() *[]*lessonResolver {
	nodes := make([]*lessonResolver, len(r.lessons))
	for i := range nodes {
		nodes[i] = &lessonResolver{Lesson: r.lessons[i], Repos: r.repos}
	}
	return &nodes
}

func (r *lessonConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	var endCursor, startCursor string
	var hasNextPage, hasPrevPage bool
	fieldName := r.pageOptions.Field.Name()
	var end, start int
	if len(r.lessons) < r.pageOptions.Limit {
		end = len(r.lessons) - 1
		hasNextPage = false
	} else if len(r.lessons) == r.pageOptions.Limit+2 {
		end = len(r.lessons) - 2
	}
	endCursor, err = r.pageOptions.Field.EncodeCursor(r.lessons[end])
	if err != nil {
		return nil, err
	}
	startCursor, err = r.pageOptions.Field.EncodeCursor(r.lessons[start])
	if err != nil {
		return nil, err
	}

	pageInfo := &pageInfoResolver{
		endCursor: endCursor,
		hasNextPage: r.pageOptions.Relation == data.GreaterThan &&
			len(r.lessons) > r.pageOptions.Limit,
		hasPrevPage: r.pageOptions.Relation == data.LessThan &&
			len(r.lessons) > r.pageOptions.Limit,
		startCursor: startCursor,
	}

	return pageInfo, nil
}

func (r *lessonConnectionResolver) TotalCount() int32 {
	return r.totalCount
}

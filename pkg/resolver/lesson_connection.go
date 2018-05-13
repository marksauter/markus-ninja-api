package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonConnectionResolver(
	lessons []*repo.LessonPermit,
	pagination *Pagination,
	totalCount int32,
	repos *repo.Repos,
) (*lessonConnectionResolver, error) {
	edges := make([]*lessonEdgeResolver, len(lessons))
	for i := range edges {
		id, err := lessons[i].ID()
		if err != nil {
			return nil, err
		}
		cursor := data.EncodeCursor(id)
		lessonEdge := NewLessonEdgeResolver(cursor, lessons[i], repos)
		edges[i] = lessonEdge
	}

	n := int32(len(edges))
	var hasNextPage, hasPreviousPage bool
	var end, start int32
	if pagination.After != nil || pagination.Before != nil {
		for i, e := range edges {
			if pagination.After != nil && e.Cursor() == pagination.After.String() {
				start = int32(i + 1)
				if i != 0 {
					hasPreviousPage = true
				} else {
					hasPreviousPage = false
				}
			}
			if pagination.Before != nil && e.Cursor() == pagination.Before.String() {
				end = int32(i - 1)
				if i != len(edges)-1 {
					hasNextPage = true
				} else {
					hasNextPage = false
				}
			}
		}
		if pagination.After == nil && int32(len(edges[:end])) > pagination.Limit() {
			start = 2
			hasNextPage = true
		} else {
			end = n - 1
			hasNextPage = false
		}
		if pagination.Before == nil && int32(len(edges[start:])) > pagination.Limit() {
			end = n - 2
			hasNextPage = true
		} else {
			end = n - 1
			hasNextPage = false
		}
	} else {
		start = 0
		hasPreviousPage = false
		if n > pagination.Limit() {
			end = n - 2
			hasNextPage = true
		} else {
			end = n - 1
			hasNextPage = false
		}
	}
	mylog.Log.Debug(start)
	endCursor := edges[end].Cursor()
	startCursor := edges[start].Cursor()

	pageInfo := &pageInfoResolver{
		endCursor:       endCursor,
		hasNextPage:     hasNextPage,
		hasPreviousPage: hasPreviousPage,
		startCursor:     startCursor,
	}

	resolver := &lessonConnectionResolver{
		edges:      edges,
		lessons:    lessons,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
		start:      start,
		end:        end,
	}
	return resolver, nil
}

type lessonConnectionResolver struct {
	edges      []*lessonEdgeResolver
	lessons    []*repo.LessonPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
	start      int32
	end        int32
}

func (r *lessonConnectionResolver) Edges() *[]*lessonEdgeResolver {
	edges := r.edges[r.start : r.end+1]
	return &edges
}

func (r *lessonConnectionResolver) Nodes() *[]*lessonResolver {
	lessons := r.lessons[r.start : r.end+1]
	nodes := make([]*lessonResolver, len(lessons))
	for i := range nodes {
		nodes[i] = &lessonResolver{Lesson: lessons[i], Repos: r.repos}
	}
	return &nodes
}

func (r *lessonConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *lessonConnectionResolver) TotalCount() int32 {
	return r.totalCount
}

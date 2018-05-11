package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonConnectionResolver(
	lessons []*repo.LessonPermit,
	pageOptions *data.PageOptions,
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

	var hasNextPage, hasPreviousPage bool
	var end, start int32
	nLessons := int32(len(lessons))
	if pageOptions.Cursor != nil {
		firstCursor := edges[0].Cursor()
		lastCursor := edges[len(edges)-1].Cursor()
		mylog.Log.Debug(*pageOptions.Cursor)
		if *pageOptions.Cursor == firstCursor || *pageOptions.Cursor == lastCursor {
			start = 0
			hasPreviousPage = false
			if nLessons > pageOptions.Limit || *pageOptions.Cursor == lastCursor {
				end = nLessons - 2
				hasNextPage = true
			} else {
				end = nLessons - 1
				hasNextPage = false
			}
		} else {
			start = 1
			hasPreviousPage = true
			if nLessons > pageOptions.Limit+1 {
				end = nLessons - 2
				hasNextPage = true
			} else {
				end = nLessons - 1
				hasNextPage = false
			}
		}
	} else {
		start = 0
		hasPreviousPage = false
		if nLessons > pageOptions.Limit {
			end = nLessons - 2
			hasNextPage = true
		} else {
			end = nLessons - 1
			hasNextPage = false
		}
	}
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

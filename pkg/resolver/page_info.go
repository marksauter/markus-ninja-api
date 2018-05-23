package resolver

import (
	"math"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type pageInfoResolver struct {
	end             int32
	endCursor       *string
	hasNextPage     bool
	hasPreviousPage bool
	start           int32
	startCursor     *string
}

type EdgeResolver interface {
	Cursor() string
}

func NewPageInfoResolver(
	edges []EdgeResolver,
	pageOptions *data.PageOptions,
) *pageInfoResolver {
	pageInfo := &pageInfoResolver{}
	n := int32(len(edges))
	if n == 0 {
		return pageInfo
	}
	var hasNextPage, hasPreviousPage bool
	end := int32(n - 1)
	start := int32(0)

	if pageOptions.After != nil || pageOptions.Before != nil {
		var after, before data.Cursor
		if pageOptions.After != nil {
			after = *pageOptions.After
		}
		if pageOptions.Before != nil {
			before = *pageOptions.Before
		}
		var haveAfterEdge, haveBeforeEdge bool
		for i, e := range edges {
			if e.Cursor() == after.String() {
				haveAfterEdge = true
				start = int32(i + 1)
				hasPreviousPage = true
			}
			if e.Cursor() == before.String() {
				haveBeforeEdge = true
				end = int32(i - 1)
				hasNextPage = true
			}
		}
		if pageOptions.After != nil {
			if !haveAfterEdge {
				start = int32(0)
				hasPreviousPage = true
			}
			if pageOptions.Before == nil {
				if pageOptions.First > 0 && n > pageOptions.Limit()+1 {
					end = int32(n - 2)
					hasNextPage = true
				} else {
					end = int32(n - 1)
					hasNextPage = false
				}
			}
		}
		if pageOptions.Before != nil {
			if !haveBeforeEdge {
				end = int32(n - 2)
				hasNextPage = true
			}
			if pageOptions.After == nil {
				if pageOptions.Last > 0 && n > pageOptions.Limit()+1 {
					start = int32(1)
					hasPreviousPage = true
				} else {
					start = int32(0)
					hasPreviousPage = false
				}
			}
		}
	} else {
		if pageOptions.First > 0 {
			start = int32(0)
			hasPreviousPage = false
			if n > pageOptions.Limit() {
				end = int32(n - 2)
				hasNextPage = true
			} else {
				end = int32(n - 1)
				hasNextPage = false
			}
		}
		if pageOptions.Last > 0 {
			end = int32(n - 1)
			hasNextPage = false
			if n > pageOptions.Limit() {
				start = int32(1)
				hasPreviousPage = true
			} else {
				start = int32(0)
				hasPreviousPage = false
			}
		}
	}

	end = int32(math.Max(0, float64(end)))
	endCursor := edges[end].Cursor()
	startCursor := edges[start].Cursor()

	pageInfo.end = end
	pageInfo.endCursor = &endCursor
	pageInfo.hasNextPage = hasNextPage
	pageInfo.hasPreviousPage = hasPreviousPage
	pageInfo.start = start
	pageInfo.startCursor = &startCursor

	return pageInfo
}

func (r *pageInfoResolver) EndCursor() *string {
	return r.endCursor
}

func (r *pageInfoResolver) HasNextPage() bool {
	return r.hasNextPage
}

func (r *pageInfoResolver) HasPreviousPage() bool {
	return r.hasPreviousPage
}

func (r *pageInfoResolver) StartCursor() *string {
	return r.startCursor
}

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
	isEmpty         bool
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
	resolver := &pageInfoResolver{}
	n := int32(len(edges))
	if n == 0 {
		return resolver
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
			if e.Cursor() == after.String {
				haveAfterEdge = true
				if n > 1 {
					start = int32(i + 1)
				} else {
					resolver.isEmpty = true
				}
				hasPreviousPage = true
			}
			if e.Cursor() == before.String {
				haveBeforeEdge = true
				if n > 1 {
					end = int32(i - 1)
				} else {
					resolver.isEmpty = true
				}
				hasNextPage = true
			}
		}
		if haveAfterEdge && haveBeforeEdge && n == 2 {
			resolver.isEmpty = true
		}
		if pageOptions.After != nil {
			if !haveAfterEdge {
				if pageOptions.Last > 0 && n == pageOptions.Limit() {
					start = int32(1)
				} else {
					start = int32(0)
				}
				hasPreviousPage = true
			}
			if pageOptions.First > 0 && n == pageOptions.Limit() {
				end = int32(n - 2)
				hasNextPage = true
			} else {
				end = int32(n - 1)
				hasNextPage = false
			}
		}
		if pageOptions.Before != nil {
			if !haveBeforeEdge {
				end = int32(n - 2)
				hasNextPage = true
			}
			if pageOptions.Last > 0 && n == pageOptions.Limit() {
				start = int32(1)
				hasPreviousPage = true
			} else {
				start = int32(0)
				hasPreviousPage = false
			}
		}
		if pageOptions.After != nil && pageOptions.Before != nil {
			if n <= pageOptions.Limit() {
				start = int32(1)
				end = int32(n - 2)
			}
		}
	} else {
		if pageOptions.First > 0 {
			start = int32(0)
			hasPreviousPage = false
			if n == pageOptions.Limit() {
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
			if n == pageOptions.Limit() {
				start = int32(1)
				hasPreviousPage = true
			} else {
				start = int32(0)
				hasPreviousPage = false
			}
		}
	}

	end = int32(math.Min(math.Max(0, float64(end)), float64(n-1)))
	endCursor := edges[end].Cursor()
	start = int32(math.Min(math.Max(0, float64(start)), float64(n-1)))
	startCursor := edges[start].Cursor()

	resolver.end = end
	resolver.endCursor = &endCursor
	resolver.hasNextPage = hasNextPage
	resolver.hasPreviousPage = hasPreviousPage
	resolver.start = start
	resolver.startCursor = &startCursor

	if end < start {
		resolver.isEmpty = true
	}

	return resolver
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

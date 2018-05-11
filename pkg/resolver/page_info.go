package resolver

type pageInfoResolver struct {
	endCursor       string
	hasNextPage     bool
	hasPreviousPage bool
	startCursor     string
}

func (r *pageInfoResolver) StartCursor() *string {
	return &r.startCursor
}

func (r *pageInfoResolver) HasNextPage() bool {
	return r.hasNextPage
}

func (r *pageInfoResolver) HasPreviousPage() bool {
	return r.hasPreviousPage
}

func (r *pageInfoResolver) EndCursor() *string {
	return &r.endCursor
}

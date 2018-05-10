package resolver

type pageInfoResolver struct {
	endCursor   string
	hasNextPage bool
	hasPrevPage bool
	startCursor string
}

func (r *pageInfoResolver) StartCursor() *string {
	return &r.startCursor
}

func (r *pageInfoResolver) HasNextPage() bool {
	return r.hasNextPage
}

func (r *pageInfoResolver) HasPrevPage() bool {
	return r.hasPrevPage
}

func (r *pageInfoResolver) EndCursor() *string {
	return &r.endCursor
}

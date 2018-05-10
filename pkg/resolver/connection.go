package resolver

type Connection interface {
	Edges() *[]*edgeResolver
	Nodes() *[]*nodeResolver
	PageInfo() *pageInfoResolver
	TotalCount() int32
}

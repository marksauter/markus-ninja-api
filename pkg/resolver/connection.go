package resolver

type Connection interface {
	Edges() *[]*edgeResolver
	Nodes() *[]*nodeResolver
	PageInfo() *pageInfoResolver
	TotalCount() int32
}

type connectionResolver struct {
	Connection
}

// func (r *connectionResolver) ToLessonConnection() (*lessonConnectionResolver, bool) {
//   resolver, ok := r.Connection.(*lessonConnectionResolver)
//   return resolver, ok
// }

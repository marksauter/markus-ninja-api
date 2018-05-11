package resolver

type Edge interface {
	Cursor() string
	Node() *nodeResolver
}

type edgeResolver struct {
	Edge
}

// func (r *edgeResolver) ToLessonEdge() (*lessonEdgeResolver, bool) {
//   resolver, ok := r.Edge.(*lessonEdgeResolver)
//   return resolver, ok
// }

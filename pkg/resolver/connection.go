package resolver

func connectionEdges(edges []interface{}, start, end int) []interface{} {
	if len(edges) > 0 {
		es := edges[start : end+1]
		return es
	}
	return edges
}

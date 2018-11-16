package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewCourseConnectionResolver(
	courses []*repo.CoursePermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.CourseFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*courseConnectionResolver, error) {
	edges := make([]*courseEdgeResolver, len(courses))
	for i := range edges {
		edge, err := NewCourseEdgeResolver(courses[i], repos, conf)
		if err != nil {
			return nil, err
		}
		edges[i] = edge
	}
	edgeResolvers := make([]EdgeResolver, len(edges))
	for i, e := range edges {
		edgeResolvers[i] = e
	}

	pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)

	resolver := &courseConnectionResolver{
		conf:     conf,
		courses:  courses,
		edges:    edges,
		filters:  filters,
		nodeID:   nodeID,
		pageInfo: pageInfo,
		repos:    repos,
	}
	return resolver, nil
}

type courseConnectionResolver struct {
	conf     *myconf.Config
	courses  []*repo.CoursePermit
	edges    []*courseEdgeResolver
	filters  *data.CourseFilterOptions
	nodeID   *mytype.OID
	pageInfo *pageInfoResolver
	repos    *repo.Repos
}

func (r *courseConnectionResolver) Edges() *[]*courseEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*courseEdgeResolver{}
}

func (r *courseConnectionResolver) Nodes() *[]*courseResolver {
	n := len(r.courses)
	nodes := make([]*courseResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		courses := r.courses[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range courses {
			nodes = append(nodes, &courseResolver{Course: s, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *courseConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *courseConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	var n int32
	if r.nodeID == nil {
		return n, nil
	}
	switch r.nodeID.Type {
	case "Study":
		return r.repos.Course().CountByStudy(ctx, r.nodeID.String, r.filters)
	case "User":
		return r.repos.Course().CountByUser(ctx, r.nodeID.String, r.filters)
	default:
		return n, errors.New("invalid node id for event total count")
	}
}

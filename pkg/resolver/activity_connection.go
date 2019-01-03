package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewActivityConnectionResolver(
	activitys []*repo.ActivityPermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.ActivityFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*activityConnectionResolver, error) {
	edges := make([]*activityEdgeResolver, len(activitys))
	for i := range edges {
		edge, err := NewActivityEdgeResolver(activitys[i], repos, conf)
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

	resolver := &activityConnectionResolver{
		conf:      conf,
		activitys: activitys,
		edges:     edges,
		filters:   filters,
		nodeID:    nodeID,
		pageInfo:  pageInfo,
		repos:     repos,
	}
	return resolver, nil
}

type activityConnectionResolver struct {
	conf      *myconf.Config
	activitys []*repo.ActivityPermit
	edges     []*activityEdgeResolver
	filters   *data.ActivityFilterOptions
	nodeID    *mytype.OID
	pageInfo  *pageInfoResolver
	repos     *repo.Repos
}

func (r *activityConnectionResolver) Edges() *[]*activityEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*activityEdgeResolver{}
}

func (r *activityConnectionResolver) Nodes() *[]*activityResolver {
	n := len(r.activitys)
	nodes := make([]*activityResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		activitys := r.activitys[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range activitys {
			nodes = append(nodes, &activityResolver{Activity: s, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *activityConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *activityConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	var n int32
	if r.nodeID == nil {
		return n, nil
	}
	switch r.nodeID.Type {
	case "Lesson":
		return r.repos.Activity().CountByLesson(ctx, r.nodeID.String, r.filters)
	case "Study":
		return r.repos.Activity().CountByStudy(ctx, r.nodeID.String, r.filters)
	case "User":
		return r.repos.Activity().CountByUser(ctx, r.nodeID.String, r.filters)
	default:
		return n, errors.New("invalid node id for activity total count")
	}
}

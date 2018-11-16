package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLabelConnectionResolver(
	labels []*repo.LabelPermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.LabelFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*labelConnectionResolver, error) {
	edges := make([]*labelEdgeResolver, len(labels))
	for i := range edges {
		edge, err := NewLabelEdgeResolver(labels[i], repos, conf)
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

	resolver := &labelConnectionResolver{
		conf:     conf,
		edges:    edges,
		filters:  filters,
		labels:   labels,
		nodeID:   nodeID,
		pageInfo: pageInfo,
		repos:    repos,
	}
	return resolver, nil
}

type labelConnectionResolver struct {
	conf     *myconf.Config
	edges    []*labelEdgeResolver
	filters  *data.LabelFilterOptions
	labels   []*repo.LabelPermit
	nodeID   *mytype.OID
	pageInfo *pageInfoResolver
	repos    *repo.Repos
}

func (r *labelConnectionResolver) Edges() *[]*labelEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*labelEdgeResolver{}
}

func (r *labelConnectionResolver) Nodes() *[]*labelResolver {
	n := len(r.labels)
	nodes := make([]*labelResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		labels := r.labels[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range labels {
			nodes = append(nodes, &labelResolver{Label: s, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *labelConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *labelConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	var n int32
	if r.nodeID == nil {
		return n, nil
	}
	switch r.nodeID.Type {
	case "Lesson":
		return r.repos.Label().CountByLabelable(ctx, r.nodeID.String, r.filters)
	case "LessonComment":
		return r.repos.Label().CountByLabelable(ctx, r.nodeID.String, r.filters)
	case "Study":
		return r.repos.Label().CountByStudy(ctx, r.nodeID.String, r.filters)
	default:
		return n, errors.New("invalid node id for event total count")
	}
}

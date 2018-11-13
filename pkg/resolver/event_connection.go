package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEventConnectionResolver(
	events []*repo.EventPermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.EventFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*eventConnectionResolver, error) {
	edges := make([]*eventEdgeResolver, len(events))
	for i := range edges {
		edge, err := NewEventEdgeResolver(events[i], repos, conf)
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

	resolver := &eventConnectionResolver{
		conf:     conf,
		edges:    edges,
		events:   events,
		filters:  filters,
		nodeID:   nodeID,
		pageInfo: pageInfo,
		repos:    repos,
	}
	return resolver, nil
}

type eventConnectionResolver struct {
	conf     *myconf.Config
	edges    []*eventEdgeResolver
	events   []*repo.EventPermit
	filters  *data.EventFilterOptions
	nodeID   *mytype.OID
	pageInfo *pageInfoResolver
	repos    *repo.Repos
}

func (r *eventConnectionResolver) Edges() *[]*eventEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*eventEdgeResolver{}
}

func (r *eventConnectionResolver) Nodes() *[]*eventResolver {
	n := len(r.events)
	nodes := make([]*eventResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		events := r.events[r.pageInfo.start : r.pageInfo.end+1]
		for _, e := range events {
			nodes = append(nodes, &eventResolver{Event: e, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *eventConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *eventConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	var n int32
	if r.nodeID == nil {
		return n, nil
	}
	switch r.nodeID.Type {
	case "Lesson":
		return r.repos.Event().CountByLesson(ctx, r.nodeID.String, r.filters)
	case "Study":
		return r.repos.Event().CountByStudy(ctx, r.nodeID.String, r.filters)
	case "User":
		return r.repos.Event().CountByUser(ctx, r.nodeID.String, r.filters)
	case "UserAsset":
		return r.repos.Event().CountByUserAsset(ctx, r.nodeID.String, r.filters)
	default:
		return n, errors.New("invalid node id for event total count")
	}
}

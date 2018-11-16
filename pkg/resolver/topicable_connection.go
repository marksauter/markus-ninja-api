package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewTopicableConnectionResolver(
	topicables []repo.NodePermit,
	pageOptions *data.PageOptions,
	topicID *mytype.OID,
	search *string,
	repos *repo.Repos,
	conf *myconf.Config,
) (*topicableConnectionResolver, error) {
	edges := make([]*topicableEdgeResolver, len(topicables))
	for i := range edges {
		edge, err := NewTopicableEdgeResolver(topicables[i], repos, conf)
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

	resolver := &topicableConnectionResolver{
		conf:       conf,
		edges:      edges,
		topicables: topicables,
		pageInfo:   pageInfo,
		repos:      repos,
		search:     search,
		topicID:    topicID,
	}
	return resolver, nil
}

type topicableConnectionResolver struct {
	conf       *myconf.Config
	edges      []*topicableEdgeResolver
	topicables []repo.NodePermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	search     *string
	topicID    *mytype.OID
}

func (r *topicableConnectionResolver) CourseCount(ctx context.Context) (int32, error) {
	filters := &data.CourseFilterOptions{
		Search: r.search,
	}
	return r.repos.Course().CountByTopic(ctx, r.topicID.String, filters)
}

func (r *topicableConnectionResolver) Edges() *[]*topicableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*topicableEdgeResolver{}
}

func (r *topicableConnectionResolver) Nodes() (*[]*topicableResolver, error) {
	n := len(r.topicables)
	nodes := make([]*topicableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		topicables := r.topicables[r.pageInfo.start : r.pageInfo.end+1]
		for _, t := range topicables {
			resolver, err := nodePermitToResolver(t, r.repos, r.conf)
			if err != nil {
				return nil, err
			}
			topicable, ok := resolver.(topicable)
			if !ok {
				return nil, errors.New("cannot convert resolver to topicable")
			}
			nodes = append(nodes, &topicableResolver{topicable})
		}
	}
	return &nodes, nil
}

func (r *topicableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *topicableConnectionResolver) StudyCount(ctx context.Context) (int32, error) {
	filters := &data.StudyFilterOptions{
		Search: r.search,
	}
	return r.repos.Study().CountByTopic(ctx, r.topicID.String, filters)
}

package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudyConnectionResolver(
	studies []*repo.StudyPermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.StudyFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*studyConnectionResolver, error) {
	edges := make([]*studyEdgeResolver, len(studies))
	for i := range edges {
		edge, err := NewStudyEdgeResolver(studies[i], repos, conf)
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

	resolver := &studyConnectionResolver{
		conf:     conf,
		edges:    edges,
		filters:  filters,
		nodeID:   nodeID,
		studies:  studies,
		pageInfo: pageInfo,
		repos:    repos,
	}
	return resolver, nil
}

type studyConnectionResolver struct {
	conf     *myconf.Config
	edges    []*studyEdgeResolver
	filters  *data.StudyFilterOptions
	nodeID   *mytype.OID
	studies  []*repo.StudyPermit
	pageInfo *pageInfoResolver
	repos    *repo.Repos
}

func (r *studyConnectionResolver) Edges() *[]*studyEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*studyEdgeResolver{}
}

func (r *studyConnectionResolver) Nodes() *[]*studyResolver {
	n := len(r.studies)
	nodes := make([]*studyResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		studies := r.studies[r.pageInfo.start : r.pageInfo.end+1]
		for _, s := range studies {
			nodes = append(nodes, &studyResolver{Study: s, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *studyConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *studyConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	var n int32
	if r.nodeID == nil {
		return n, nil
	}
	switch r.nodeID.Type {
	case "User":
		return r.repos.Study().CountByUser(ctx, r.nodeID.String, r.filters)
	default:
		return n, errors.New("invalid node id for study total count")
	}
}

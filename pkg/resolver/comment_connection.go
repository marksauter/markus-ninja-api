package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewCommentConnectionResolver(
	comments []*repo.CommentPermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.CommentFilterOptions,
	repos *repo.Repos,
	conf *myconf.Config,
) (*commentConnectionResolver, error) {
	edges := make([]*commentEdgeResolver, len(comments))
	for i := range edges {
		edge, err := NewCommentEdgeResolver(comments[i], repos, conf)
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

	resolver := &commentConnectionResolver{
		conf:     conf,
		edges:    edges,
		filters:  filters,
		comments: comments,
		nodeID:   nodeID,
		pageInfo: pageInfo,
		repos:    repos,
	}
	return resolver, nil
}

type commentConnectionResolver struct {
	conf     *myconf.Config
	edges    []*commentEdgeResolver
	filters  *data.CommentFilterOptions
	comments []*repo.CommentPermit
	nodeID   *mytype.OID
	pageInfo *pageInfoResolver
	repos    *repo.Repos
}

func (r *commentConnectionResolver) Edges() *[]*commentEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*commentEdgeResolver{}
}

func (r *commentConnectionResolver) Nodes() *[]*commentResolver {
	n := len(r.comments)
	nodes := make([]*commentResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		comments := r.comments[r.pageInfo.start : r.pageInfo.end+1]
		for _, l := range comments {
			nodes = append(nodes, &commentResolver{Comment: l, Conf: r.conf, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *commentConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *commentConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	switch r.nodeID.Type {
	case "Lesson":
		return r.repos.Comment().CountByCommentable(ctx, r.nodeID.String, r.filters)
	case "Study":
		return r.repos.Comment().CountByStudy(ctx, r.nodeID.String, r.filters)
	case "User":
		return r.repos.Comment().CountByUser(ctx, r.nodeID.String, r.filters)
	case "UserAsset":
		return r.repos.Comment().CountByCommentable(ctx, r.nodeID.String, r.filters)
	default:
		var n int32
		return n, errors.New("invalid node id for comment total count")
	}
}

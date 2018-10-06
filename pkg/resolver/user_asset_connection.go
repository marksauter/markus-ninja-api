package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserAssetConnectionResolver(
	repos *repo.Repos,
	userAssets []*repo.UserAssetPermit,
	pageOptions *data.PageOptions,
	nodeID *mytype.OID,
	filters *data.UserAssetFilterOptions,
) (*userAssetConnectionResolver, error) {
	edges := make([]*userAssetEdgeResolver, len(userAssets))
	for i := range edges {
		edge, err := NewUserAssetEdgeResolver(userAssets[i], repos)
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

	resolver := &userAssetConnectionResolver{
		edges:      edges,
		filters:    filters,
		pageInfo:   pageInfo,
		repos:      repos,
		userAssets: userAssets,
		nodeID:     nodeID,
	}
	return resolver, nil
}

type userAssetConnectionResolver struct {
	edges      []*userAssetEdgeResolver
	filters    *data.UserAssetFilterOptions
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	userAssets []*repo.UserAssetPermit
	nodeID     *mytype.OID
}

func (r *userAssetConnectionResolver) Edges() *[]*userAssetEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*userAssetEdgeResolver{}
}

func (r *userAssetConnectionResolver) Nodes() *[]*userAssetResolver {
	n := len(r.userAssets)
	nodes := make([]*userAssetResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		userAssets := r.userAssets[r.pageInfo.start : r.pageInfo.end+1]
		for _, l := range userAssets {
			nodes = append(nodes, &userAssetResolver{UserAsset: l, Repos: r.repos})
		}
	}
	return &nodes
}

func (r *userAssetConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

func (r *userAssetConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	switch r.nodeID.Type {
	case "Study":
		return r.repos.UserAsset().CountByStudy(ctx, r.nodeID.String, r.filters)
	case "User":
		return r.repos.UserAsset().CountByUser(ctx, r.nodeID.String, r.filters)
	default:
		var n int32
		return n, errors.New("invalid node ID for user asset total count")
	}
}

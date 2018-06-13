package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserAssetConnectionResolver(
	userAssets []*repo.UserAssetPermit,
	pageOptions *data.PageOptions,
	totalCount int32,
	repos *repo.Repos,
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
		userAssets: userAssets,
		pageInfo:   pageInfo,
		repos:      repos,
		totalCount: totalCount,
	}
	return resolver, nil
}

type userAssetConnectionResolver struct {
	edges      []*userAssetEdgeResolver
	userAssets []*repo.UserAssetPermit
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	totalCount int32
}

func (r *userAssetConnectionResolver) Edges() *[]*userAssetEdgeResolver {
	if len(r.edges) > 0 {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &r.edges
}

func (r *userAssetConnectionResolver) Nodes() *[]*userAssetResolver {
	n := len(r.userAssets)
	nodes := make([]*userAssetResolver, 0, n)
	if n > 0 {
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

func (r *userAssetConnectionResolver) TotalCount() int32 {
	return r.totalCount
}

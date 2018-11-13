package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLabelableConnectionResolver(
	labelables []repo.NodePermit, pageOptions *data.PageOptions,
	labelID *mytype.OID,
	search *string,
	repos *repo.Repos,
	conf *myconf.Config,
) (*labelableConnectionResolver, error) {
	edges := make([]*labelableEdgeResolver, len(labelables))
	for i := range edges {
		edge, err := NewLabelableEdgeResolver(labelables[i], repos, conf)
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

	resolver := &labelableConnectionResolver{
		conf:       conf,
		edges:      edges,
		labelables: labelables,
		labelID:    labelID,
		pageInfo:   pageInfo,
		repos:      repos,
		search:     search,
	}
	return resolver, nil
}

type labelableConnectionResolver struct {
	conf       *myconf.Config
	edges      []*labelableEdgeResolver
	labelables []repo.NodePermit
	labelID    *mytype.OID
	pageInfo   *pageInfoResolver
	repos      *repo.Repos
	search     *string
}

func (r *labelableConnectionResolver) Edges() *[]*labelableEdgeResolver {
	if len(r.edges) > 0 && !r.pageInfo.isEmpty {
		edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
		return &edges
	}
	return &[]*labelableEdgeResolver{}
}

func (r *labelableConnectionResolver) LessonCount(ctx context.Context) (int32, error) {
	filters := &data.LessonFilterOptions{
		Search: r.search,
	}
	return r.repos.Lesson().CountByLabel(ctx, r.labelID.String, filters)
}

func (r *labelableConnectionResolver) LessonCommentCount(ctx context.Context) (int32, error) {
	// filters := &data.LessonCommentFilterOptions{
	//   Search: r.search,
	// }
	// return r.repos.LessonComment().CountByLabel(ctx, r.labelID.String)
	return int32(0), nil
}

func (r *labelableConnectionResolver) Nodes() (*[]*labelableResolver, error) {
	n := len(r.labelables)
	nodes := make([]*labelableResolver, 0, n)
	if n > 0 && !r.pageInfo.isEmpty {
		labelables := r.labelables[r.pageInfo.start : r.pageInfo.end+1]
		for _, t := range labelables {
			resolver, err := nodePermitToResolver(t, r.repos, r.conf)
			if err != nil {
				return nil, err
			}
			labelable, ok := resolver.(labelable)
			if !ok {
				return nil, errors.New("cannot convert resolver to labelable")
			}
			nodes = append(nodes, &labelableResolver{labelable})
		}
	}
	return &nodes, nil
}

func (r *labelableConnectionResolver) PageInfo() (*pageInfoResolver, error) {
	return r.pageInfo, nil
}

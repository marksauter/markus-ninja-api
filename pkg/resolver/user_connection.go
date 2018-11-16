package resolver

// import (
//   "errors"
//
//   "github.com/marksauter/markus-ninja-api/pkg/data"
//   "github.com/marksauter/markus-ninja-api/pkg/myconf"
//   "github.com/marksauter/markus-ninja-api/pkg/mytype"
//   "github.com/marksauter/markus-ninja-api/pkg/repo"
// )
//
// func NewUserConnectionResolver(
//   users []*repo.UserPermit,
//   pageOptions *data.PageOptions,
//   nodeID *mytype.OID,
//   filters *data.UserFilterOptions,
//   repos *repo.Repos,
//   conf *myconf.Config,
// ) (*userConnectionResolver, error) {
//   edges := make([]*userEdgeResolver, len(users))
//   for i := range edges {
//     id, err := users[i].ID()
//     if err != nil {
//       return nil, err
//     }
//     cursor := data.EncodeCursor(id.String)
//     edge := NewUserEdgeResolver(cursor, users[i], repos, conf)
//     edges[i] = edge
//   }
//   edgeResolvers := make([]EdgeResolver, len(edges))
//   for i, e := range edges {
//     edgeResolvers[i] = e
//   }
//
//   pageInfo := NewPageInfoResolver(edgeResolvers, pageOptions)
//
//   resolver := &userConnectionResolver{
//     conf:     conf,
//     edges:    edges,
//     filters:  filters,
//     nodeID:   nodeID,
//     pageInfo: pageInfo,
//     repos:    repos,
//     users:    users,
//   }
//   return resolver, nil
// }
//
// type userConnectionResolver struct {
//   conf     *myconf.Config
//   edges    []*userEdgeResolver
//   filters  *data.UserFilterOptions
//   nodeID   *mytype.OID
//   pageInfo *pageInfoResolver
//   repos    *repo.Repos
//   users    []*repo.UserPermit
// }
//
// func (r *userConnectionResolver) Edges() *[]*userEdgeResolver {
//   if len(r.edges) > 0 && !r.pageInfo.isEmpty {
//     edges := r.edges[r.pageInfo.start : r.pageInfo.end+1]
//     return &edges
//   }
//   return &[]*userEdgeResolver{}
// }
//
// func (r *userConnectionResolver) Nodes() *[]*userResolver {
//   n := len(r.users)
//   nodes := make([]*userResolver, 0, n)
//   if n > 0 && !r.pageInfo.isEmpty {
//     users := r.users[r.pageInfo.start : r.pageInfo.end+1]
//     for _, s := range users {
//       nodes = append(nodes, &userResolver{User: s, Conf: r.conf, Repos: r.repos})
//     }
//   }
//   return &nodes
// }
//
// func (r *userConnectionResolver) PageInfo() (*pageInfoResolver, error) {
//   return r.pageInfo, nil
// }
//
// func (r *userConnectionResolver) TotalCount() int32 {
//   var n int32
//   if r.nodeID == nil {
//     return n, nil
//   }
//   switch r.nodeID.Type {
//   case "Lesson":
//     return r.repos.Event().CountByLesson(ctx, r.nodeID.String, r.filters)
//   case "Study":
//     return r.repos.Event().CountByStudy(ctx, r.nodeID.String, r.filters)
//   case "User":
//     return r.repos.Event().CountByUser(ctx, r.nodeID.String, r.filters)
//   case "UserAsset":
//     return r.repos.Event().CountByUserAsset(ctx, r.nodeID.String, r.filters)
//   default:
//     return n, errors.New("invalid node id for event total count")
//   }
// }

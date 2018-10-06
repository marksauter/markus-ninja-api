package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func (r *RootResolver) Asset(
	ctx context.Context,
	args struct {
		Owner string
		Study string
		Name  string
	},
) (*userAssetResolver, error) {
	userAsset, err := r.Repos.UserAsset().GetByUserStudyAndName(
		ctx,
		args.Owner,
		args.Study,
		args.Name,
	)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Repos: r.Repos}, nil
}

func (r *RootResolver) Node(
	ctx context.Context,
	args struct{ ID string },
) (*nodeResolver, error) {
	id, err := mytype.ParseOID(args.ID)
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetNode(ctx, id)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	node, ok := resolver.(node)
	if !ok {
		return nil, errors.New("cannot convert resolver to node")
	}
	return &nodeResolver{node}, nil
}

func (r *RootResolver) Nodes(ctx context.Context, args struct {
	IDs []string
}) ([]*nodeResolver, error) {
	nodes := make([]*nodeResolver, len(args.IDs))
	for i, id := range args.IDs {
		nodeID, err := mytype.ParseOID(id)
		if err != nil {
			return nil, err
		}
		permit, err := r.Repos.GetNode(ctx, nodeID)
		if err != nil {
			return nil, err
		}
		resolver, err := nodePermitToResolver(permit, r.Repos)
		if err != nil {
			return nil, err
		}
		node, ok := resolver.(node)
		if !ok {
			return nil, errors.New("cannot convert resolver to node")
		}
		nodes[i] = &nodeResolver{node}
	}
	return nodes, nil
}

func (r *RootResolver) Relay() *RootResolver {
	return r
}

func (r *RootResolver) Search(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
		Query   string
		Type    string
	},
) (*searchableConnectionResolver, error) {
	searchType, err := ParseSearchType(args.Type)
	if err != nil {
		return nil, err
	}
	searchOrder, err := ParseSearchOrder(searchType, args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		searchOrder,
	)
	if err != nil {
		return nil, err
	}

	permits := []repo.NodePermit{}

	switch searchType {
	case SearchTypeCourse:
		courses, err := r.Repos.Course().Search(ctx, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(courses))
		for i, l := range courses {
			permits[i] = l
		}
	case SearchTypeLabel:
		labels, err := r.Repos.Label().Search(ctx, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(labels))
		for i, l := range labels {
			permits[i] = l
		}
	case SearchTypeLesson:
		lessons, err := r.Repos.Lesson().Search(ctx, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(lessons))
		for i, l := range lessons {
			permits[i] = l
		}
	case SearchTypeStudy:
		studies, err := r.Repos.Study().Search(ctx, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(studies))
		for i, l := range studies {
			permits[i] = l
		}
	case SearchTypeTopic:
		topics, err := r.Repos.Topic().Search(ctx, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(topics))
		for i, l := range topics {
			permits[i] = l
		}
	case SearchTypeUser:
		users, err := r.Repos.User().Search(ctx, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(users))
		for i, l := range users {
			permits[i] = l
		}
	case SearchTypeUserAsset:
		userAssets, err := r.Repos.UserAsset().Search(ctx, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(userAssets))
		for i, l := range userAssets {
			permits[i] = l
		}
	}

	return NewSearchableConnectionResolver(
		r.Repos,
		permits,
		pageOptions,
		args.Query,
	)
}

func (r *RootResolver) Study(
	ctx context.Context,
	args struct {
		Name  string
		Owner string
	},
) (*studyResolver, error) {
	study, err := r.Repos.Study().GetByUserAndName(ctx, args.Owner, args.Name)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *RootResolver) Topic(
	ctx context.Context,
	args struct {
		Name  string
		Owner string
	},
) (*topicResolver, error) {
	topic, err := r.Repos.Topic().GetByName(ctx, args.Name)
	if err != nil {
		return nil, err
	}
	return &topicResolver{Topic: topic, Repos: r.Repos}, nil
}

func (r *RootResolver) User(ctx context.Context, args struct {
	Login string
}) (*userResolver, error) {
	if args.Login == repo.Guest {
		return nil, nil
	}
	user, err := r.Repos.User().GetByLogin(ctx, args.Login)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *RootResolver) Viewer(ctx context.Context) (*userResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	if viewer.Login.String == repo.Guest {
		return nil, nil
	}
	user, err := r.Repos.User().Get(ctx, viewer.ID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

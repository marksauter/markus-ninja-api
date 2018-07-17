package resolver

import (
	"context"
	"errors"
	"fmt"

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
	args struct{ Id string },
) (*nodeResolver, error) {
	id, err := mytype.ParseOID(args.Id)
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
	Ids []string
}) ([]*nodeResolver, error) {
	nodes := make([]*nodeResolver, len(args.Ids))
	for i, id := range args.Ids {
		nodeId, err := mytype.ParseOID(id)
		if err != nil {
			return nil, err
		}
		permit, err := r.Repos.GetNode(ctx, nodeId)
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
		Within  *string
	},
) (*searchResultItemConnectionResolver, error) {
	var within *mytype.OID
	if args.Within != nil {
		var err error
		within, err = mytype.ParseOID(*args.Within)
		if err != nil {
			return nil, err
		}
		if within.Type != "User" && within.Type != "Study" {
			return nil, fmt.Errorf("cannot search within %s", within.Type)
		}
	}
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

	lessonCount, err := r.Repos.Lesson().CountBySearch(ctx, within, args.Query)
	if err != nil {
		return nil, err
	}
	studyCount, err := r.Repos.Study().CountBySearch(ctx, within, args.Query)
	if err != nil {
		return nil, err
	}
	topicCount, err := r.Repos.Topic().CountBySearch(ctx, within, args.Query)
	if err != nil {
		return nil, err
	}
	userCount, err := r.Repos.User().CountBySearch(ctx, args.Query)
	if err != nil {
		return nil, err
	}
	userAssetCount, err := r.Repos.UserAsset().CountBySearch(ctx, within, args.Query)
	if err != nil {
		return nil, err
	}
	counts := &resultItemCounts{
		Lesson:    lessonCount,
		Study:     studyCount,
		Topic:     topicCount,
		User:      userCount,
		UserAsset: userAssetCount,
	}
	permits := []repo.NodePermit{}

	switch searchType {
	case SearchTypeLesson:
		lessons, err := r.Repos.Lesson().Search(ctx, within, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(lessons))
		for i, l := range lessons {
			permits[i] = l
		}
	case SearchTypeStudy:
		studies, err := r.Repos.Study().Search(ctx, within, args.Query, pageOptions)
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
		userAssets, err := r.Repos.UserAsset().Search(ctx, within, args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.NodePermit, len(userAssets))
		for i, l := range userAssets {
			permits[i] = l
		}
	}

	return NewSearchResultItemConnectionResolver(
		r.Repos,
		permits,
		pageOptions,
		counts,
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
	user, err := r.Repos.User().Get(ctx, viewer.Id.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

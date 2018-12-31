package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
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
	return &userAssetResolver{UserAsset: userAsset, Conf: r.Conf, Repos: r.Repos}, nil
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
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
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
		resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
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
	resolver := searchableConnectionResolver{}
	searchType, err := ParseSearchType(args.Type)
	if err != nil {
		return &resolver, err
	}
	searchOrder, err := ParseSearchOrder(searchType, args.OrderBy)
	if err != nil {
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		searchOrder,
	)
	if err != nil {
		return &resolver, err
	}

	permits := []repo.NodePermit{}

	switch searchType {
	case SearchTypeActivity:
		filters := &data.ActivityFilterOptions{
			Search: &args.Query,
		}
		activities, err := r.Repos.Activity().Search(ctx, pageOptions, filters)
		if err != nil {
			return &resolver, err
		}
		permits = make([]repo.NodePermit, len(activities))
		for i, l := range activities {
			permits[i] = l
		}
	case SearchTypeCourse:
		filters := &data.CourseFilterOptions{
			IsPublished: util.NewBool(true),
			Search:      &args.Query,
		}
		courses, err := r.Repos.Course().Search(ctx, pageOptions, filters)
		if err != nil {
			return &resolver, err
		}
		permits = make([]repo.NodePermit, len(courses))
		for i, l := range courses {
			permits[i] = l
		}
	case SearchTypeLabel:
		filters := &data.LabelFilterOptions{
			Search: &args.Query,
		}
		labels, err := r.Repos.Label().Search(ctx, pageOptions, filters)
		if err != nil {
			return &resolver, err
		}
		permits = make([]repo.NodePermit, len(labels))
		for i, l := range labels {
			permits[i] = l
		}
	case SearchTypeLesson:
		filters := &data.LessonFilterOptions{
			IsPublished: util.NewBool(true),
			Search:      &args.Query,
		}
		lessons, err := r.Repos.Lesson().Search(ctx, pageOptions, filters)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return &resolver, err
		}
		permits = make([]repo.NodePermit, len(lessons))
		for i, l := range lessons {
			permits[i] = l
		}
	case SearchTypeStudy:
		filters := &data.StudyFilterOptions{
			Search: &args.Query,
		}
		studies, err := r.Repos.Study().Search(ctx, pageOptions, filters)
		if err != nil {
			return &resolver, err
		}
		permits = make([]repo.NodePermit, len(studies))
		for i, l := range studies {
			permits[i] = l
		}
	case SearchTypeTopic:
		filters := &data.TopicFilterOptions{
			Search: &args.Query,
		}
		topics, err := r.Repos.Topic().Search(ctx, pageOptions, filters)
		if err != nil {
			return &resolver, err
		}
		permits = make([]repo.NodePermit, len(topics))
		for i, l := range topics {
			permits[i] = l
		}
	case SearchTypeUser:
		filters := &data.UserFilterOptions{
			Search: &args.Query,
		}
		users, err := r.Repos.User().Search(ctx, pageOptions, filters)
		if err != nil {
			return &resolver, err
		}
		permits = make([]repo.NodePermit, len(users))
		for i, l := range users {
			permits[i] = l
		}
	case SearchTypeUserAsset:
		filters := &data.UserAssetFilterOptions{
			Search: &args.Query,
		}
		userAssets, err := r.Repos.UserAsset().Search(ctx, pageOptions, filters)
		if err != nil {
			return &resolver, err
		}
		permits = make([]repo.NodePermit, len(userAssets))
		for i, l := range userAssets {
			permits[i] = l
		}
	}

	return NewSearchableConnectionResolver(
		permits,
		pageOptions,
		args.Query,
		r.Repos,
		r.Conf,
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
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
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
	return &topicResolver{Topic: topic, Conf: r.Conf, Repos: r.Repos}, nil
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
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
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
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

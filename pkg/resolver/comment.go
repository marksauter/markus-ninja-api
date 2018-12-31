package resolver

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type commentResolver struct {
	Conf    *myconf.Config
	Comment *repo.CommentPermit
	Repos   *repo.Repos
}

func (r *commentResolver) Author(ctx context.Context) (*userResolver, error) {
	userID, err := r.Comment.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *commentResolver) Body() (string, error) {
	body, err := r.Comment.Body()
	if err != nil {
		return "", err
	}
	return body.String, nil
}

func (r *commentResolver) BodyHTML(ctx context.Context) (mygql.HTML, error) {
	body, err := r.Comment.Body()
	if err != nil {
		return "", err
	}
	return mygql.HTML(body.ToHTML()), nil
}

func (r *commentResolver) BodyText() (string, error) {
	body, err := r.Comment.Body()
	if err != nil {
		return "", err
	}
	return body.ToText(), nil
}

func (r *commentResolver) Commentable(ctx context.Context) (*commentableResolver, error) {
	commentableID, err := r.Comment.CommentableID()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetCommentable(ctx, commentableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	commentable, ok := resolver.(commentable)
	if !ok {
		return nil, errors.New("cannot convert resolver to commentable")
	}
	return &commentableResolver{commentable}, nil
}

func (r *commentResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Comment.CreatedAt()
	return graphql.Time{t}, err
}

func (r *commentResolver) Draft() (string, error) {
	return r.Comment.Draft()
}

func (r *commentResolver) DraftBackup(
	ctx context.Context,
	args struct{ ID string },
) (*commentDraftBackupResolver, error) {
	commentID, err := r.Comment.ID()
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseInt(args.ID, 10, 32)
	if err != nil {
		return nil, errors.New("invalid backup id")
	}

	draftBackup, err := r.Repos.CommentDraftBackup().Get(ctx, commentID.String, int32(id))
	if err != nil {
		return nil, err
	}
	return &commentDraftBackupResolver{
		Conf:               r.Conf,
		CommentDraftBackup: draftBackup,
		Repos:              r.Repos,
	}, nil
}

func (r *commentResolver) DraftBackups(
	ctx context.Context,
) ([]*commentDraftBackupResolver, error) {
	resolvers := []*commentDraftBackupResolver{}

	commentID, err := r.Comment.ID()
	if err != nil {
		return resolvers, err
	}

	draftBackups, err := r.Repos.CommentDraftBackup().GetByComment(ctx, commentID.String)
	if err != nil {
		return resolvers, err
	}

	resolvers = make([]*commentDraftBackupResolver, len(draftBackups))
	for i, b := range draftBackups {
		resolvers[i] = &commentDraftBackupResolver{
			Conf:               r.Conf,
			CommentDraftBackup: b,
			Repos:              r.Repos,
		}
	}

	return resolvers, nil
}

func (r *commentResolver) ID() (graphql.ID, error) {
	id, err := r.Comment.ID()
	return graphql.ID(id.String), err
}

func (r *commentResolver) IsPublished() (bool, error) {
	return r.Comment.IsPublished()
}

func (r *commentResolver) Labels(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.LabelFilterOptions
		First    *int32
		Last     *int32
		OrderBy  *OrderArg
	},
) (*labelConnectionResolver, error) {
	commentID, err := r.Comment.ID()
	if err != nil {
		return nil, err
	}
	labelOrder, err := ParseLabelOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		labelOrder,
	)
	if err != nil {
		return nil, err
	}

	labels, err := r.Repos.Label().GetByLabelable(
		ctx,
		commentID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	labelConnectionResolver, err := NewLabelConnectionResolver(
		labels,
		pageOptions,
		commentID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return labelConnectionResolver, nil
}

func (r *commentResolver) LastEditedAt() (graphql.Time, error) {
	t, err := r.Comment.LastEditedAt()
	return graphql.Time{t}, err
}

func (r *commentResolver) PublishedAt() (*graphql.Time, error) {
	t, err := r.Comment.PublishedAt()
	if err != nil {
		return nil, err
	}
	return &graphql.Time{t}, nil
}

func (r *commentResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI

	commentType, err := r.Type()
	if err != nil {
		return uri, err
	}

	commentable, err := r.Commentable(ctx)
	if err != nil {
		return uri, err
	}
	var basePath mygql.URI
	switch commentType {
	case mytype.CommentableTypeLesson.String():
		lesson, ok := commentable.ToLesson()
		if !ok {
			return uri, errors.New("cannot convert commentable to lesson")
		}
		basePath, err = lesson.ResourcePath(ctx)
		if err != nil {
			return uri, err
		}
	case mytype.CommentableTypeUserAsset.String():
		userAsset, ok := commentable.ToUserAsset()
		if !ok {
			return uri, errors.New("cannot convert commentable to user asset")
		}
		basePath, err = userAsset.ResourcePath(ctx)
		if err != nil {
			return uri, err
		}
	}

	createdAt, err := r.Comment.CreatedAt()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf(
		"%s#comment%d",
		string(basePath),
		createdAt.Unix(),
	))
	return uri, nil
}

func (r *commentResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Comment.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *commentResolver) Type() (string, error) {
	return r.Comment.Type()
}

func (r *commentResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Comment.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *commentResolver) URL(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", r.Conf.ClientURL, resourcePath))
	return uri, nil
}

func (r *commentResolver) ViewerCanDelete(ctx context.Context) bool {
	comment := r.Comment.Get()
	return r.Repos.Comment().ViewerCanDelete(ctx, comment)
}

func (r *commentResolver) ViewerCanUpdate(ctx context.Context) bool {
	comment := r.Comment.Get()
	return r.Repos.Comment().ViewerCanUpdate(ctx, comment)
}

func (r *commentResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userID, err := r.Comment.UserID()
	if err != nil {
		return false, err
	}

	return viewer.ID.String == userID.String, nil
}

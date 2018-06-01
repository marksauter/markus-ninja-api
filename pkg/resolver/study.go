package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Study = studyResolver

type studyResolver struct {
	Study *repo.StudyPermit
	Repos *repo.Repos
}

func (r *studyResolver) AdvancedAt() (*graphql.Time, error) {
	t, err := r.Study.AdvancedAt()
	if err != nil {
		return nil, err
	}
	if t != nil {
		return &graphql.Time{*t}, nil
	}
	return nil, nil
}

func (r *studyResolver) Asset(
	ctx context.Context,
	args struct{ Name string },
) (*userAssetResolver, error) {
	id, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	userAsset, err := r.Repos.UserAsset().GetByStudyIdAndName(id.String, args.Name)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Repos: r.Repos}, nil
}

func (r *studyResolver) Assets(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *UserAssetOrderArg
	},
) (*userAssetConnectionResolver, error) {
	id, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	userAssetOrder, err := ParseUserAssetOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		userAssetOrder,
	)
	if err != nil {
		return nil, err
	}

	userAssets, err := r.Repos.UserAsset().GetByStudyId(id, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.UserAsset().CountByStudy(id.String)
	if err != nil {
		return nil, err
	}
	userAssetConnectionResolver, err := NewUserAssetConnectionResolver(
		userAssets,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return userAssetConnectionResolver, nil
}

func (r *studyResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Study.CreatedAt()
	return graphql.Time{t}, err
}

func (r *studyResolver) Description() (string, error) {
	return r.Study.Description()
}

func (r *studyResolver) DescriptionHTML() (mygql.HTML, error) {
	description, err := r.Description()
	if err != nil {
		return "", err
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", description))
	return h, nil
}

func (r *studyResolver) ID() (graphql.ID, error) {
	id, err := r.Study.ID()
	return graphql.ID(id.String), err
}

func (r *studyResolver) Lesson(
	ctx context.Context,
	args struct{ Number int32 },
) (*lessonResolver, error) {
	id, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().GetByStudyNumber(id.String, args.Number)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}

func (r *studyResolver) Lessons(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *LessonOrderArg
	},
) (*lessonConnectionResolver, error) {
	id, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lessonOrder, err := ParseLessonOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		lessonOrder,
	)
	if err != nil {
		return nil, err
	}

	lessons, err := r.Repos.Lesson().GetByStudyId(id.String, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Lesson().CountByStudy(id.String)
	if err != nil {
		return nil, err
	}
	lessonConnectionResolver, err := NewLessonConnectionResolver(
		lessons,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return lessonConnectionResolver, nil
}

func (r *studyResolver) LessonComments(
	ctx context.Context,
	args struct {
		After  *string
		Before *string
		First  *int32
		Last   *int32
	},
) (*lessonCommentConnectionResolver, error) {
	id, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lessonCommentOrder := NewLessonCommentOrder(data.DESC, LessonCommentCreatedAt)

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		lessonCommentOrder,
	)
	if err != nil {
		return nil, err
	}

	lessonComments, err := r.Repos.LessonComment().GetByStudyId(id.String, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.LessonComment().CountByStudy(id.String)
	if err != nil {
		return nil, err
	}
	lessonCommentConnectionResolver, err := NewLessonCommentConnectionResolver(
		lessonComments,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return lessonCommentConnectionResolver, nil
}

func (r *studyResolver) LessonCount() (int32, error) {
	id, err := r.Study.ID()
	if err != nil {
		var count int32
		return count, err
	}
	return r.Repos.Lesson().CountByStudy(id.String)
}

func (r *studyResolver) Name() (string, error) {
	return r.Study.Name()
}

func (r *studyResolver) NameWithOwner() (string, error) {
	name, err := r.Name()
	if err != nil {
		return "", err
	}
	owner, err := r.Owner()
	if err != nil {
		return "", err
	}
	ownerLogin, err := owner.Login()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", ownerLogin, name), nil
}

func (r *studyResolver) Owner() (*userResolver, error) {
	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *studyResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	nameWithOwner, err := r.NameWithOwner()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("/%s", nameWithOwner))
	return uri, nil
}

func (r *studyResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Study.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *studyResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *studyResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userId, err := r.Study.UserId()
	if err != nil {
		return false, err
	}

	return viewer.Id.String == userId.String, nil
}

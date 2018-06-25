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
	"github.com/marksauter/markus-ninja-api/pkg/util"
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

func (r *studyResolver) AppleGivers(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*appleGiverConnectionResolver, error) {
	appleOrder, err := ParseAppleOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		appleOrder,
	)
	if err != nil {
		return nil, err
	}

	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	users, err := r.Repos.User().GetByApple(
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.User().CountByApple(studyId.String)
	if err != nil {
		return nil, err
	}
	appleGiverConnectionResolver, err := NewAppleGiverConnectionResolver(
		users,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return appleGiverConnectionResolver, nil
}

func (r *studyResolver) Asset(
	ctx context.Context,
	args struct{ Name string },
) (*userAssetResolver, error) {
	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	userAsset, err := r.Repos.UserAsset().GetByName(
		userId.String,
		studyId.String,
		args.Name,
	)
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
		OrderBy *OrderArg
	},
) (*userAssetConnectionResolver, error) {
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

	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	userAssets, err := r.Repos.UserAsset().GetByStudy(
		userId,
		studyId,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.UserAsset().CountByStudy(
		userId.String,
		studyId.String,
	)
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
	descriptionHTML := util.MarkdownToHTML([]byte(description))
	gqlHTML := mygql.HTML(descriptionHTML)
	return gqlHTML, nil
}

func (r *studyResolver) ID() (graphql.ID, error) {
	id, err := r.Study.ID()
	return graphql.ID(id.String), err
}

func (r *studyResolver) Lesson(
	ctx context.Context,
	args struct{ Number int32 },
) (*lessonResolver, error) {
	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().GetByNumber(
		userId.String,
		studyId.String,
		args.Number,
	)
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
		OrderBy *OrderArg
	},
) (*lessonConnectionResolver, error) {
	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	studyId, err := r.Study.ID()
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

	lessons, err := r.Repos.Lesson().GetByStudy(
		userId.String,
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Lesson().CountByStudy(
		userId.String,
		studyId.String,
	)
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
	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	studyId, err := r.Study.ID()
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

	lessonComments, err := r.Repos.LessonComment().GetByStudy(
		userId.String,
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.LessonComment().CountByStudy(
		userId.String,
		studyId.String,
	)
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
	userId, err := r.Study.UserId()
	if err != nil {
		var count int32
		return count, err
	}
	studyId, err := r.Study.ID()
	if err != nil {
		var count int32
		return count, err
	}
	return r.Repos.Lesson().CountByStudy(
		userId.String,
		studyId.String,
	)
}

func (r *studyResolver) Name() (string, error) {
	return r.Study.Name()
}

func (r *studyResolver) NameWithOwner() (string, error) {
	name, err := r.Name()
	if err != nil {
		return "", err
	}
	ownerLogin, err := r.Study.UserLogin()
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

func (r *studyResolver) Students(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*studentConnectionResolver, error) {
	studentOrder, err := ParseStudentOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		studentOrder,
	)
	if err != nil {
		return nil, err
	}

	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	users, err := r.Repos.User().GetStudents(
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.User().CountByEnrollable(studyId.String)
	if err != nil {
		return nil, err
	}
	studentConnectionResolver, err := NewStudentConnectionResolver(
		users,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return studentConnectionResolver, nil
}

func (r *studyResolver) Topics(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*topicConnectionResolver, error) {
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	topicOrder, err := ParseTopicOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		topicOrder,
	)
	if err != nil {
		return nil, err
	}

	topics, err := r.Repos.Topic().GetByStudy(
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Topic().CountByStudy(studyId.String)
	if err != nil {
		return nil, err
	}
	topicConnectionResolver, err := NewTopicConnectionResolver(
		topics,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return topicConnectionResolver, nil
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
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, resourcePath))
	return uri, nil
}

func (r *studyResolver) ViewerCanUpdate() bool {
	study := r.Study.Get()
	return r.Repos.Study().ViewerCanUpdate(study)
}

func (r *studyResolver) ViewerHasAppled(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return false, err
	}

	if _, err := r.Repos.StudyApple().Get(studyId.String, viewer.Id.String); err != nil {
		if err == data.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *studyResolver) ViewerHasEnrolled(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return false, err
	}

	if _, err := r.Repos.StudyEnroll().Get(studyId.String, viewer.Id.String); err != nil {
		if err == data.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

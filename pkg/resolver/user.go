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

type User = userResolver

type userResolver struct {
	User  *repo.UserPermit
	Repos *repo.Repos
}

func (r *userResolver) Appled(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
		Type    string
	},
) (*appleableConnectionResolver, error) {
	appleableType, err := ParseAppleableType(args.Type)
	if err != nil {
		return nil, err
	}
	appleableOrder, err := ParseAppleableOrder(appleableType, args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		appleableOrder,
	)
	if err != nil {
		return nil, err
	}

	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}

	studyCount, err := r.Repos.Study().CountByApplee(ctx, id.String)
	if err != nil {
		return nil, err
	}
	permits := []repo.Permit{}

	switch appleableType {
	case AppleableTypeStudy:
		studies, err := r.Repos.Study().GetByApplee(ctx, id.String, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.Permit, len(studies))
		for i, l := range studies {
			permits[i] = l
		}
	default:
		return nil, fmt.Errorf("invalid type %s for appleable type", appleableType.String())
	}

	return NewAppleableConnectionResolver(
		r.Repos,
		permits,
		pageOptions,
		studyCount,
	)
}

func (r *userResolver) Assets(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*userAssetConnectionResolver, error) {
	id, err := r.User.ID()
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

	userAssets, err := r.Repos.UserAsset().GetByUser(ctx, id, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.UserAsset().CountByUser(ctx, id.String)
	if err != nil {
		return nil, err
	}
	resolver, err := NewUserAssetConnectionResolver(
		userAssets,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) Bio() (string, error) {
	return r.User.Bio()
}

func (r *userResolver) BioHTML() (mygql.HTML, error) {
	bio, err := r.Bio()
	if err != nil {
		return "", err
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", bio))
	return h, nil
}

func (r *userResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.User.CreatedAt()
	return graphql.Time{t}, err
}

func (r *userResolver) Email() (string, error) {
	return r.User.PublicEmail()
}

func (r *userResolver) Emails(
	ctx context.Context,
	args struct {
		After  *string
		Before *string
		First  *int32
		Last   *int32
	},
) (*emailConnectionResolver, error) {
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}

	order := NewEmailOrder(data.ASC, EmailType)
	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		order,
	)
	if err != nil {
		return nil, err
	}

	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	emails, err := r.Repos.Email().GetByUser(ctx, id, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Email().CountByUser(ctx, id.String)
	if err != nil {
		return nil, err
	}
	resolver, err := NewEmailConnectionResolver(
		emails,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) Enrolled(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
		Type    string
	},
) (*enrollableConnectionResolver, error) {
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}
	enrollableType, err := ParseEnrollableType(args.Type)
	if err != nil {
		return nil, err
	}
	enrollableOrder, err := ParseEnrollableOrder(enrollableType, args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		enrollableOrder,
	)
	if err != nil {
		return nil, err
	}

	lessonCount, err := r.Repos.Lesson().CountByEnrollee(ctx, id.String)
	if err != nil {
		return nil, err
	}
	studyCount, err := r.Repos.Study().CountByEnrollee(ctx, id.String)
	if err != nil {
		return nil, err
	}
	userCount, err := r.Repos.User().CountByEnrollee(ctx, id.String)
	if err != nil {
		return nil, err
	}
	permits := []repo.Permit{}

	switch enrollableType {
	case EnrollableTypeLesson:
		lessons, err := r.Repos.Lesson().GetByEnrollee(ctx, id.String, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.Permit, len(lessons))
		for i, l := range lessons {
			permits[i] = l
		}
	case EnrollableTypeStudy:
		studies, err := r.Repos.Study().GetByEnrollee(ctx, id.String, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.Permit, len(studies))
		for i, s := range studies {
			permits[i] = s
		}
	case EnrollableTypeUser:
		users, err := r.Repos.User().GetByEnrollee(ctx, id.String, pageOptions)
		if err != nil {
			return nil, err
		}
		permits = make([]repo.Permit, len(users))
		for i, u := range users {
			permits[i] = u
		}
	}
	return NewEnrollableConnectionResolver(
		r.Repos,
		permits,
		pageOptions,
		lessonCount,
		studyCount,
		userCount,
	)
}

func (r *userResolver) Enrollees(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*enrolleeConnectionResolver, error) {
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}

	enrolleeOrder, err := ParseEnrolleeOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		enrolleeOrder,
	)
	if err != nil {
		return nil, err
	}

	users, err := r.Repos.User().GetEnrollees(
		ctx,
		id.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.User().CountByEnrollable(ctx, id.String)
	if err != nil {
		return nil, err
	}
	resolver, err := NewEnrolleeConnectionResolver(
		users,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) Notifications(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*notificationConnectionResolver, error) {
	userId, err := r.User.ID()
	if err != nil {
		return nil, err
	}
	notificationOrder, err := ParseNotificationOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		notificationOrder,
	)
	if err != nil {
		return nil, err
	}

	notifications, err := r.Repos.Notification().GetByUser(
		ctx,
		userId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Notification().CountByUser(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	notificationConnectionResolver, err := NewNotificationConnectionResolver(
		notifications,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return notificationConnectionResolver, nil
}

func (r *userResolver) ID() (graphql.ID, error) {
	id, err := r.User.ID()
	return graphql.ID(id.String), err
}

func (r *userResolver) IsSiteAdmin() bool {
	for _, role := range r.User.Roles() {
		if role == data.AdminRole {
			return true
		}
	}
	return false
}

func (r *userResolver) IsViewer(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	id, err := r.User.ID()
	if err != nil {
		return false, err
	}
	return viewer.Id.String == id.String, nil
}

func (r *userResolver) Lessons(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*lessonConnectionResolver, error) {
	id, err := r.User.ID()
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

	lessons, err := r.Repos.Lesson().GetByUser(ctx, id.String, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Lesson().CountByUser(ctx, id.String)
	if err != nil {
		return nil, err
	}
	resolver, err := NewLessonConnectionResolver(
		lessons,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) Login() (string, error) {
	return r.User.Login()
}

func (r *userResolver) Name() (string, error) {
	return r.User.Name()
}

func (r *userResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	login, err := r.User.Login()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("/%s", login))
	return uri, nil
}

func (r *userResolver) Study(
	ctx context.Context,
	args struct{ Name string },
) (*studyResolver, error) {
	userId, err := r.User.ID()
	if err != nil {
		return nil, err
	}

	study, err := r.Repos.Study().GetByName(ctx, userId.String, args.Name)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *userResolver) Studies(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*studyConnectionResolver, error) {
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}
	studyOrder, err := ParseStudyOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		studyOrder,
	)
	if err != nil {
		return nil, err
	}

	studies, err := r.Repos.Study().GetByUser(ctx, id.String, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Study().CountByUser(ctx, id.String)
	if err != nil {
		return nil, err
	}
	resolver, err := NewStudyConnectionResolver(
		studies,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.User.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *userResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	login, err := r.User.Login()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, login))
	return uri, nil
}

func (r *userResolver) ViewerHasEnrolled(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userId, err := r.User.ID()
	if err != nil {
		return false, err
	}

	enrolled := &data.Enrolled{}
	enrolled.EnrollableId.Set(userId)
	enrolled.UserId.Set(viewer.Id)
	if _, err := r.Repos.Enrolled().Get(ctx, enrolled); err != nil {
		if err == data.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

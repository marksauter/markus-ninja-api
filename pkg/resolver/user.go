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
	},
) (*appledStudyConnectionResolver, error) {
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}
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

	studies, err := r.Repos.Study().GetByAppled(id.String, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Study().CountByAppled(id.String)
	if err != nil {
		return nil, err
	}
	appledStudyConnectionResolver, err := NewAppledStudyConnectionResolver(
		studies,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return appledStudyConnectionResolver, nil
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

	userAssets, err := r.Repos.UserAsset().GetByUser(id, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.UserAsset().CountByUser(id.String)
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

func (r *userResolver) Bio() (string, error) {
	return r.User.Profile()
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

	emails, err := r.Repos.Email().GetByUser(id, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Email().CountByUser(id.String)
	if err != nil {
		return nil, err
	}
	emailConnectionResolver, err := NewEmailConnectionResolver(
		emails,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return emailConnectionResolver, nil
}

func (r *userResolver) Enrolled(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*enrolledStudyConnectionResolver, error) {
	return nil, nil
}

func (r *userResolver) Pupils(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*pupilConnectionResolver, error) {
	return nil, nil
}

func (r *userResolver) Tutors(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*tutorConnectionResolver, error) {
	return nil, nil
}

func (r *userResolver) ID() (graphql.ID, error) {
	id, err := r.User.ID()
	return graphql.ID(id.String), err
}

func (r *userResolver) IsSiteAdmin() bool {
	for _, role := range r.User.Roles() {
		if role == "ADMIN" {
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

	lessons, err := r.Repos.Lesson().GetByUser(id.String, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Lesson().CountByUser(id.String)
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

	study, err := r.Repos.Study().GetByName(userId.String, args.Name)
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

	studies, err := r.Repos.Study().GetByUser(id.String, pageOptions)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Study().CountByUser(id.String)
	if err != nil {
		return nil, err
	}
	studyConnectionResolver, err := NewStudyConnectionResolver(
		studies,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return studyConnectionResolver, nil
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

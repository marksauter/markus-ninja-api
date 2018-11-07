package resolver

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type lessonCommentResolver struct {
	LessonComment *repo.LessonCommentPermit
	Repos         *repo.Repos
}

func (r *lessonCommentResolver) Author(ctx context.Context) (*userResolver, error) {
	userID, err := r.LessonComment.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) Body() (string, error) {
	body, err := r.LessonComment.Body()
	if err != nil {
		return "", err
	}
	return body.String, nil
}

func (r *lessonCommentResolver) BodyHTML(ctx context.Context) (mygql.HTML, error) {
	commentBody, err := r.LessonComment.Body()
	if err != nil {
		return "", err
	}
	body := *commentBody
	studyID, err := r.LessonComment.StudyID()
	if err != nil {
		return "", err
	}
	bodyStr := body.String

	study, err := r.Study(ctx)
	if err != nil {
		return "", err
	}
	studyPath, err := study.ResourcePath(ctx)
	if err != nil {
		return "", err
	}
	lessonNumberRefToLink := func(s string) string {
		result := mytype.NumberRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		number := result[1]
		n, err := strconv.ParseInt(number, 10, 32)
		if err != nil {
			return s
		}
		exists, err := r.Repos.Lesson().ExistsByNumber(ctx, studyID.String, int32(n))
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[#%[3]d](%[1]s%[2]s/lesson/%[3]d)",
			clientURL,
			studyPath,
			n,
		))
	}
	bodyStr = mytype.NumberRefRegexp.ReplaceAllStringFunc(bodyStr, lessonNumberRefToLink)

	crossStudyRefToLink := func(s string) string {
		result := mytype.CrossStudyRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		owner := result[1]
		name := result[2]
		number := result[3]
		n, err := strconv.ParseInt(number, 10, 32)
		if err != nil {
			return s
		}
		exists, err := r.Repos.Lesson().ExistsByOwnerStudyAndNumber(ctx, owner, name, int32(n))
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[%[2]s/%[3]s#%[4]d](%[1]s/%[2]s/%[3]s/lesson/%[4]d)",
			clientURL,
			owner,
			name,
			n,
		))
	}
	bodyStr = mytype.CrossStudyRefRegexp.ReplaceAllStringFunc(bodyStr, crossStudyRefToLink)

	userRefToLink := func(s string) string {
		result := mytype.AtRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		name := result[1]
		exists, err := r.Repos.User().ExistsByLogin(ctx, name)
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[@%[2]s](%[1]s/%[2]s)",
			clientURL,
			name,
		))
	}
	bodyStr = mytype.AtRefRegexp.ReplaceAllStringFunc(bodyStr, userRefToLink)

	if err := body.Set(bodyStr); err != nil {
		return "", err
	}

	return mygql.HTML(body.ToHTML()), nil
}

func (r *lessonCommentResolver) BodyText() (string, error) {
	body, err := r.LessonComment.Body()
	if err != nil {
		return "", err
	}
	return body.ToText(), nil
}

func (r *lessonCommentResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.LessonComment.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) Draft() (string, error) {
	return r.LessonComment.Draft()
}

func (r *lessonCommentResolver) ID() (graphql.ID, error) {
	id, err := r.LessonComment.ID()
	return graphql.ID(id.String), err
}

func (r *lessonCommentResolver) IsPublished() (bool, error) {
	return r.LessonComment.IsPublished()
}

func (r *lessonCommentResolver) Labels(
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
	lessonCommentID, err := r.LessonComment.ID()
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
		lessonCommentID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Label().CountByLabelable(ctx, lessonCommentID.String, args.FilterBy)
	if err != nil {
		return nil, err
	}
	labelConnectionResolver, err := NewLabelConnectionResolver(
		r.Repos,
		labels,
		pageOptions,
		count,
	)
	if err != nil {
		return nil, err
	}
	return labelConnectionResolver, nil
}

func (r *lessonCommentResolver) LastEditedAt() (graphql.Time, error) {
	t, err := r.LessonComment.LastEditedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) Lesson(ctx context.Context) (*lessonResolver, error) {
	lessonID, err := r.LessonComment.LessonID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().Get(ctx, lessonID.String)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) PublishedAt() (*graphql.Time, error) {
	t, err := r.LessonComment.PublishedAt()
	if err != nil {
		return nil, err
	}
	return &graphql.Time{t}, nil
}

func (r *lessonCommentResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	lesson, err := r.Lesson(ctx)
	if err != nil {
		return uri, err
	}
	lessonPath, err := lesson.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	createdAt, err := r.LessonComment.CreatedAt()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf(
		"%s#lesson-comment%d",
		string(lessonPath),
		createdAt.Unix(),
	))
	return uri, nil
}

func (r *lessonCommentResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.LessonComment.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.LessonComment.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) URL(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *lessonCommentResolver) ViewerCanDelete(ctx context.Context) bool {
	lessonComment := r.LessonComment.Get()
	return r.Repos.LessonComment().ViewerCanDelete(ctx, lessonComment)
}

func (r *lessonCommentResolver) ViewerCanUpdate(ctx context.Context) bool {
	lessonComment := r.LessonComment.Get()
	return r.Repos.LessonComment().ViewerCanUpdate(ctx, lessonComment)
}

func (r *lessonCommentResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userID, err := r.LessonComment.UserID()
	if err != nil {
		return false, err
	}

	return viewer.ID.String == userID.String, nil
}

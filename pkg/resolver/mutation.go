package resolver

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/jackc/pgx/pgtype"
	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type AddCourseLessonInput struct {
	CourseID string
	LessonID string
}

func (r *RootResolver) AddCourseLesson(
	ctx context.Context,
	args struct{ Input AddCourseLessonInput },
) (*addCourseLessonPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		err := errors.New("viewer not found")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	courseLesson := &data.CourseLesson{}
	if err := courseLesson.CourseID.Set(args.Input.CourseID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("invalid value for courseId")
	}
	if err := courseLesson.LessonID.Set(args.Input.LessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("invalid value for lessonId")
	}

	lesson, err := r.Repos.Lesson().Pull(ctx, args.Input.LessonID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("lesson not found")
	}
	isPublished, err := lesson.IsPublished()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}
	if !isPublished {
		err := errors.New("lesson not published")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	studyID, err := lesson.StudyID()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}

	_, err = r.Repos.CourseLesson().Connect(ctx, courseLesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	eventPayload, err := data.NewLessonAddedToCoursePayload(
		&courseLesson.LessonID,
		&courseLesson.CourseID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	event, err := data.NewLessonEvent(eventPayload, studyID, &viewer.ID, true)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.Repos.Event().Create(ctx, event); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	return &addCourseLessonPayloadResolver{
		Conf:     r.Conf,
		CourseID: &courseLesson.CourseID,
		LessonID: &courseLesson.LessonID,
		Repos:    r.Repos,
	}, nil
}

type AddEmailInput struct {
	Email string
}

func (r *RootResolver) AddEmail(
	ctx context.Context,
	args struct{ Input AddEmailInput },
) (*addEmailPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	email := &data.Email{}
	if err := email.Value.Set(args.Input.Email); err != nil {
		return nil, errors.New("Invalid email")
	}
	email.UserID.Set(&viewer.ID)
	if err := email.UserID.Set(&viewer.ID); err != nil {
		return nil, errors.New("Invalid userId")
	}

	emailPermit, err := r.Repos.Email().Create(ctx, email)
	if err != nil {
		return nil, err
	}

	evt := &data.EVT{}
	if err := evt.EmailID.Set(&email.ID); err != nil {
		mylog.Log.WithError(err).Error("failed to set evt email_id")
		return nil, myerr.SomethingWentWrongError
	}
	if err := evt.UserID.Set(&viewer.ID); err != nil {
		mylog.Log.WithError(err).Error("failed to set evt user_id")
		return nil, myerr.SomethingWentWrongError
	}

	evtPermit, err := r.Repos.EVT().Create(ctx, evt)
	if err != nil {
		return nil, err
	}

	resolver := &addEmailPayloadResolver{
		Conf:  r.Conf,
		Email: emailPermit,
		EVT:   evtPermit,
		Repos: r.Repos,
	}
	sendMailInput := &service.SendEmailVerificationMailInput{
		EmailID:   email.ID.Short,
		To:        args.Input.Email,
		UserLogin: viewer.Login.String,
		Token:     evt.Token.String,
	}
	if err := r.Svcs.Mail.SendEmailVerificationMail(sendMailInput); err != nil {
		return resolver, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return resolver, err
		}
	}

	return resolver, nil
}

type AddLabelInput struct {
	LabelID     string
	LabelableID string
}

func (r *RootResolver) AddLabel(
	ctx context.Context,
	args struct{ Input AddLabelInput },
) (*addLabelPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	labeled := &data.Labeled{}
	if err := labeled.LabelID.Set(args.Input.LabelID); err != nil {
		return nil, errors.New("invalid labeled label_id")
	}
	if err := labeled.LabelableID.Set(args.Input.LabelableID); err != nil {
		return nil, errors.New("invalid labeled labelable_id")
	}

	_, err = r.Repos.Labeled().Connect(ctx, labeled)
	if err != nil {
		return nil, err
	}

	if labeled.LabelableID.Type == mytype.LessonNodeType.String() {
		lesson, err := r.Repos.Lesson().Pull(ctx, args.Input.LabelableID)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, errors.New("lesson not found")
		}
		isPublished, err := lesson.IsPublished()
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, myerr.SomethingWentWrongError
		}
		studyID, err := lesson.StudyID()
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, myerr.SomethingWentWrongError
		}

		eventPayload, err := data.NewLessonLabeledPayload(
			&labeled.LabelableID,
			&labeled.LabelID,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		isPublic := isPublished
		event, err := data.NewLessonEvent(eventPayload, studyID, &viewer.ID, isPublic)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		if _, err := r.Repos.Event().Create(ctx, event); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &addLabelPayloadResolver{
		Conf:        r.Conf,
		LabelID:     &labeled.LabelID,
		LabelableID: &labeled.LabelableID,
		Repos:       r.Repos,
	}, nil
}

type AddLessonCommentInput struct {
	LessonCommentID string
}

func (r *RootResolver) AddLessonComment(
	ctx context.Context,
	args struct{ Input AddLessonCommentInput },
) (*addLessonCommentPayloadResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	currentLessonCommentPermit, err := r.Repos.LessonComment().Get(ctx, args.Input.LessonCommentID)
	if err != nil {
		return nil, errors.New("lesson comment not found")
	}
	draft, err := currentLessonCommentPermit.Draft()
	if err != nil {
		return nil, err
	}

	lessonComment := &data.LessonComment{}
	if err := lessonComment.ID.Set(args.Input.LessonCommentID); err != nil {
		return nil, errors.New("Invalid lessonCommentId")
	}
	if err := lessonComment.Body.Set(draft); err != nil {
		mylog.Log.WithError(err).Error("failed to set lesson comment's body to its draft")
		return nil, myerr.SomethingWentWrongError
	}
	if err := lessonComment.PublishedAt.Set(time.Now()); err != nil {
		mylog.Log.WithError(err).Error("failed to set lesson comment published_at")
		return nil, myerr.SomethingWentWrongError
	}

	lessonCommentPermit, err := r.Repos.LessonComment().Update(ctx, lessonComment)
	if err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &addLessonCommentPayloadResolver{
		Conf:          r.Conf,
		LessonComment: lessonCommentPermit,
		Repos:         r.Repos,
	}, nil
}

type CreateCourseInput struct {
	Description *string
	Name        string
	StudyID     string
}

func (r *RootResolver) CreateCourse(
	ctx context.Context,
	args struct{ Input CreateCourseInput },
) (*createCoursePayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	course := &data.Course{}
	if err := course.Description.Set(args.Input.Description); err != nil {
		return nil, errors.New("invalid course description")
	}
	if err := course.Name.Set(args.Input.Name); err != nil {
		return nil, errors.New("invalid course name")
	}
	if err := course.StudyID.Set(args.Input.StudyID); err != nil {
		return nil, errors.New("invalid course study_id")
	}
	if err := course.UserID.Set(&viewer.ID); err != nil {
		return nil, errors.New("invalid course user_id")
	}

	coursePermit, err := r.Repos.Course().Create(ctx, course)
	if err != nil {
		return nil, err
	}

	return &createCoursePayloadResolver{
		Conf:    r.Conf,
		Course:  coursePermit,
		StudyID: &course.StudyID,
		Repos:   r.Repos,
	}, nil
}

type CreateLabelInput struct {
	Color       string
	Description *string
	Name        string
	StudyID     string
}

func (r *RootResolver) CreateLabel(
	ctx context.Context,
	args struct{ Input CreateLabelInput },
) (*createLabelPayloadResolver, error) {
	label := &data.Label{}
	if err := label.Color.Set(args.Input.Color); err != nil {
		return nil, errors.New("Invalid color")
	}
	if err := label.Description.Set(args.Input.Description); err != nil {
		return nil, errors.New("Invalid description")
	}
	if err := label.Name.Set(args.Input.Name); err != nil {
		return nil, errors.New("Invalid name")
	}
	if err := label.StudyID.Set(args.Input.StudyID); err != nil {
		return nil, errors.New("Invalid studyId")
	}
	labelPermit, err := r.Repos.Label().Create(ctx, label)
	if err != nil {
		return nil, err
	}

	return &createLabelPayloadResolver{
		Conf:    r.Conf,
		Label:   labelPermit,
		StudyID: &label.StudyID,
		Repos:   r.Repos,
	}, nil
}

type CreateLessonInput struct {
	StudyID string
	Title   string
}

func (r *RootResolver) CreateLesson(
	ctx context.Context,
	args struct{ Input CreateLessonInput },
) (*createLessonPayloadResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	lesson := &data.Lesson{}
	if err := lesson.StudyID.Set(args.Input.StudyID); err != nil {
		return nil, errors.New("Invalid studyId")
	}
	if err := lesson.Title.Set(args.Input.Title); err != nil {
		return nil, errors.New("Invalid title")
	}
	if err := lesson.UserID.Set(&viewer.ID); err != nil {
		mylog.Log.WithError(err).Error("failed to set lesson user_id")
		return nil, myerr.SomethingWentWrongError
	}

	lessonPermit, err := r.Repos.Lesson().Create(ctx, lesson)
	if err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &createLessonPayloadResolver{
		Conf:    r.Conf,
		Lesson:  lessonPermit,
		StudyID: &lesson.StudyID,
		Repos:   r.Repos,
	}, nil
}

type CreateStudyInput struct {
	Description *string
	Name        string
}

func (r *RootResolver) CreateStudy(
	ctx context.Context,
	args struct{ Input CreateStudyInput },
) (*createStudyPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	study := &data.Study{}
	if err := study.Description.Set(args.Input.Description); err != nil {
		return nil, errors.New("Invalid description")
	}
	if err := study.Name.Set(args.Input.Name); err != nil {
		return nil, errors.New("Invalid name")
	}
	if err := study.UserID.Set(&viewer.ID); err != nil {
		mylog.Log.Error("failed to set study user_id")
		return nil, myerr.SomethingWentWrongError
	}

	studyPermit, err := r.Repos.Study().Create(ctx, study)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create study")
		return nil, myerr.SomethingWentWrongError
	}
	study = studyPermit.Get()

	return &createStudyPayloadResolver{
		Conf:   r.Conf,
		Study:  studyPermit,
		UserID: &study.UserID,
		Repos:  r.Repos,
	}, nil
}

type CreateUserInput struct {
	Email    string
	Login    string
	Password string
}

func (r *RootResolver) CreateUser(
	ctx context.Context,
	args struct{ Input CreateUserInput },
) (*userResolver, error) {
	user := &data.User{}
	if err := user.PrimaryEmail.Set(args.Input.Email); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("Invalid email")
	}
	if err := user.Login.Set(args.Input.Login); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("Invalid login")
	}
	if err := user.Password.Set(args.Input.Password); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("Invalid password")
	}

	ok, err := user.Login.IsBlacklisted()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	} else if ok {
		err := errors.New("Username unavailable")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	userPermit, err := r.Repos.User().Create(ctx, user)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return &userResolver{
		Conf:  r.Conf,
		Repos: r.Repos,
		User:  userPermit,
	}, nil
}

type CreateUserAssetInput struct {
	AssetID     string
	Description *string
	Name        string
	StudyID     string
}

func (r *RootResolver) CreateUserAsset(
	ctx context.Context,
	args struct{ Input CreateUserAssetInput },
) (*createUserAssetPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	assetID, err := strconv.ParseInt(args.Input.AssetID, 10, 64)
	if err != nil {
		return nil, err
	}

	userAsset := &data.UserAsset{}
	if err := userAsset.AssetID.Set(assetID); err != nil {
		return nil, errors.New("Invalid assetId")
	}
	if err := userAsset.Description.Set(args.Input.Description); err != nil {
		return nil, errors.New("Invalid description")
	}
	if err := userAsset.Name.Set(args.Input.Name); err != nil {
		return nil, errors.New("Invalid name")
	}
	if err := userAsset.StudyID.Set(args.Input.StudyID); err != nil {
		return nil, errors.New("Invalid studyId")
	}
	if err := userAsset.UserID.Set(&viewer.ID); err != nil {
		return nil, errors.New("Invalid userId")
	}

	userAssetPermit, err := r.Repos.UserAsset().Create(ctx, userAsset)
	if err != nil {
		return nil, err
	}

	return &createUserAssetPayloadResolver{
		Conf:      r.Conf,
		Repos:     r.Repos,
		StudyID:   &userAsset.StudyID,
		UserAsset: userAssetPermit,
		UserID:    &userAsset.UserID,
	}, nil
}

type DeleteEmailInput struct {
	EmailID string
}

func (r *RootResolver) DeleteEmail(
	ctx context.Context,
	args struct{ Input DeleteEmailInput },
) (*deleteEmailPayloadResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	emailPermit, err := r.Repos.Email().Get(ctx, args.Input.EmailID)
	if err != nil {
		return nil, err
	}

	email := emailPermit.Get()

	isVerified := true
	filters := &data.EmailFilterOptions{
		IsVerified: &isVerified,
	}
	n, err := r.Repos.Email().CountByUser(
		ctx,
		email.UserID.String,
		filters,
	)
	if err != nil {
		return nil, err
	}
	if n == 1 {
		return nil, errors.New("cannot delete your only verified email")
	}

	if err := r.Repos.Email().Delete(ctx, email); err != nil {
		return nil, err
	}

	if email.Type.V == mytype.PrimaryEmail {
		var newPrimaryEmail *data.Email
		emails, err := r.Repos.Email().GetByUser(
			ctx,
			email.UserID.String,
			nil,
			filters,
		)
		if err != nil {
			return nil, err
		}
		n := len(emails)
		for i, email := range emails {
			e := email.Get()
			if e.Type.V == mytype.BackupEmail {
				newPrimaryEmail = e
			}
			if newPrimaryEmail == nil && i == n-1 {
				newPrimaryEmail = e
			}
		}
		newPrimaryEmail.Type.Set(mytype.PrimaryEmail)
		if _, err := r.Repos.Email().Update(ctx, newPrimaryEmail); err != nil {
			return nil, err
		}
	}

	resolver := &deleteEmailPayloadResolver{
		Conf:    r.Conf,
		EmailID: &email.ID,
		UserID:  &email.UserID,
		Repos:   r.Repos,
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return resolver, err
		}
	}

	return resolver, nil
}

type DeleteLabelInput struct {
	LabelID string
}

func (r *RootResolver) DeleteLabel(
	ctx context.Context,
	args struct{ Input DeleteLabelInput },
) (*deleteLabelPayloadResolver, error) {
	labelPermit, err := r.Repos.Label().Get(ctx, args.Input.LabelID)
	if err != nil {
		return nil, err
	}
	label := labelPermit.Get()

	if err := r.Repos.Label().Delete(ctx, label); err != nil {
		return nil, err
	}

	return &deleteLabelPayloadResolver{
		Conf:    r.Conf,
		LabelID: &label.ID,
		Repos:   r.Repos,
		StudyID: &label.StudyID,
	}, nil
}

type DeleteLessonInput struct {
	LessonID string
}

func (r *RootResolver) DeleteLesson(
	ctx context.Context,
	args struct{ Input DeleteLessonInput },
) (*deleteLessonPayloadResolver, error) {
	lessonPermit, err := r.Repos.Lesson().Get(ctx, args.Input.LessonID)
	if err != nil {
		return nil, err
	}
	lesson := lessonPermit.Get()

	if err := r.Repos.Lesson().Delete(ctx, lesson); err != nil {
		return nil, err
	}

	return &deleteLessonPayloadResolver{
		Conf:     r.Conf,
		LessonID: &lesson.ID,
		Repos:    r.Repos,
		StudyID:  &lesson.StudyID,
	}, nil
}

type DeleteLessonCommentInput struct {
	LessonCommentID string
}

func (r *RootResolver) DeleteLessonComment(
	ctx context.Context,
	args struct{ Input DeleteLessonCommentInput },
) (*deleteLessonCommentPayloadResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	lessonCommentPermit, err := r.Repos.LessonComment().Get(ctx, args.Input.LessonCommentID)
	if err != nil {
		return nil, err
	}
	lessonComment := lessonCommentPermit.Get()

	if err := r.Repos.LessonComment().Delete(ctx, lessonComment); err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &deleteLessonCommentPayloadResolver{
		Conf:            r.Conf,
		LessonCommentID: &lessonComment.ID,
		LessonID:        &lessonComment.LessonID,
		Repos:           r.Repos,
	}, nil
}

type DeleteCourseInput struct {
	CourseID string
}

func (r *RootResolver) DeleteCourse(
	ctx context.Context,
	args struct{ Input DeleteCourseInput },
) (*deleteCoursePayloadResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	coursePermit, err := r.Repos.Course().Get(ctx, args.Input.CourseID)
	if err != nil {
		return nil, err
	}
	course := coursePermit.Get()

	if err := r.Repos.Course().Delete(ctx, course); err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &deleteCoursePayloadResolver{
		Conf:     r.Conf,
		CourseID: &course.ID,
		Repos:    r.Repos,
		StudyID:  &course.StudyID,
	}, nil
}

type DeleteStudyInput struct {
	StudyID string
}

func (r *RootResolver) DeleteStudy(
	ctx context.Context,
	args struct{ Input DeleteStudyInput },
) (*deleteStudyPayloadResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	studyPermit, err := r.Repos.Study().Get(ctx, args.Input.StudyID)
	if err != nil {
		return nil, err
	}
	study := studyPermit.Get()

	if err := r.Repos.Study().Delete(ctx, study); err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &deleteStudyPayloadResolver{
		Conf:    r.Conf,
		OwnerID: &study.UserID,
		Repos:   r.Repos,
		StudyID: &study.ID,
	}, nil
}

type DeleteUserAssetInput struct {
	UserAssetID string
}

func (r *RootResolver) DeleteUserAsset(
	ctx context.Context,
	args struct{ Input DeleteUserAssetInput },
) (*deleteUserAssetPayloadResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	userAssetPermit, err := r.Repos.UserAsset().Get(ctx, args.Input.UserAssetID)
	if err != nil {
		return nil, err
	}
	userAsset := userAssetPermit.Get()

	if err := r.Repos.UserAsset().Delete(ctx, userAsset); err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &deleteUserAssetPayloadResolver{
		Conf:        r.Conf,
		Repos:       r.Repos,
		StudyID:     &userAsset.StudyID,
		UserAssetID: &userAsset.ID,
	}, nil
}

type DeleteViewerAccountInput struct {
	Login    string
	Password string
}

func (r *RootResolver) DeleteViewerAccount(
	ctx context.Context,
	args struct{ Input DeleteViewerAccountInput },
) (*deleteViewerAccountPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	if viewer.Login.String != args.Input.Login {
		return nil, errors.New("invalid credentials")
	}

	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}

	user, err := data.GetUserCredentialsByLogin(db, args.Input.Login)
	if err != nil {
		return nil, err
	}

	if err := user.Password.CompareToPassword(args.Input.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := r.Repos.User().Delete(ctx, user); err != nil {
		return nil, err
	}

	return &deleteViewerAccountPayloadResolver{
		Conf:     r.Conf,
		Repos:    r.Repos,
		ViewerID: &user.ID,
	}, nil
}

type GiveAppleInput struct {
	AppleableID string
}

func (r *RootResolver) GiveApple(
	ctx context.Context,
	args struct{ Input GiveAppleInput },
) (*appleableResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	appled := &data.Appled{}
	if err := appled.AppleableID.Set(args.Input.AppleableID); err != nil {
		return nil, errors.New("invalid appleable id")
	}
	if err := appled.UserID.Set(&viewer.ID); err != nil {
		return nil, errors.New("invalid appleable user_id")
	}
	_, err = r.Repos.Appled().Connect(ctx, appled)
	if err != nil {
		return nil, err
	}
	appleablePermit, err := r.Repos.GetAppleable(ctx, &appled.AppleableID)
	if err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	resolver, err := nodePermitToResolver(appleablePermit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	appleable, ok := resolver.(appleable)
	if !ok {
		return nil, errors.New("cannot convert resolver to appleable")
	}
	return &appleableResolver{appleable}, nil
}

var InvalidCredentialsError = errors.New("invalid credentials")

type LoginUserInput struct {
	Login    string
	Password string
}

func (r *RootResolver) LoginUser(
	ctx context.Context,
	args struct{ Input LoginUserInput },
) (*loginUserPayloadResolver, error) {
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	var user *data.User
	if err := checkmail.ValidateFormat(args.Input.Login); err != nil {
		user, err = data.GetUserCredentialsByLogin(db, args.Input.Login)
		if err != nil {
			return nil, InvalidCredentialsError
		}
	} else {
		user, err = data.GetUserCredentialsByEmail(db, args.Input.Login)
		if err != nil {
			return nil, InvalidCredentialsError
		}
	}

	if err := user.Password.CompareToPassword(args.Input.Password); err != nil {
		return nil, InvalidCredentialsError
	}

	exp := time.Now().Add(time.Hour * time.Duration(24)).Unix()
	payload := myjwt.Payload{Exp: exp, Iat: time.Now().Unix(), Sub: user.ID.String}
	jwt, err := r.Svcs.Auth.SignJWT(&payload)
	if err != nil {
		return nil, InternalServerError
	}

	return &loginUserPayloadResolver{
		AccessToken: jwt,
		Conf:        r.Conf,
		Viewer:      user,
		Repos:       r.Repos,
	}, nil
}

func (r *RootResolver) LogoutUser(
	ctx context.Context,
) (*logoutUserPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	return &logoutUserPayloadResolver{
		UserID: &viewer.ID,
		Repos:  r.Repos,
	}, nil
}

type MarkNotificationAsReadInput struct {
	NotificationID string
}

func (r *RootResolver) MarkNotificationAsRead(
	ctx context.Context,
	args struct{ Input MarkNotificationAsReadInput },
) (*graphql.ID, error) {
	notification := &data.Notification{}
	if err := notification.ID.Set(args.Input.NotificationID); err != nil {
		return nil, errors.New("invalid notification id")
	}

	err := r.Repos.Notification().Delete(ctx, notification)
	if err != nil {
		return nil, err
	}

	id := graphql.ID(notification.ID.String)
	return &id, nil
}

func (r *RootResolver) MarkAllNotificationsAsRead(
	ctx context.Context,
) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}

	notification := &data.Notification{}
	if err := notification.UserID.Set(&viewer.ID); err != nil {
		return false, errors.New("invalid notification user_id")
	}

	if err := r.Repos.Notification().DeleteByUser(ctx, notification); err != nil {
		return false, err
	}
	return true, nil
}

type MarkAllStudyNotificationAsReadInput struct {
	StudyID string
}

func (r *RootResolver) MarkAllStudyNotificationsAsRead(
	ctx context.Context,
	args struct {
		Input MarkAllStudyNotificationAsReadInput
	},
) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}

	notification := &data.Notification{}
	if err := notification.StudyID.Set(args.Input.StudyID); err != nil {
		return false, errors.New("invalid notification study_id")
	}
	if err := notification.UserID.Set(&viewer.ID); err != nil {
		return false, errors.New("invalid notification user_id")
	}

	if err := r.Repos.Notification().DeleteByStudy(ctx, notification); err != nil {
		return false, err
	}
	return true, nil
}

type MoveCourseLessonInput struct {
	AfterLessonID string
	CourseID      string
	LessonID      string
}

func (r *RootResolver) MoveCourseLesson(
	ctx context.Context,
	args struct{ Input MoveCourseLessonInput },
) (*moveCourseLessonPayloadResolver, error) {
	courseLesson := &data.CourseLesson{}
	if err := courseLesson.CourseID.Set(args.Input.CourseID); err != nil {
		return nil, errors.New("invalid value for courseId")
	}
	if err := courseLesson.LessonID.Set(args.Input.LessonID); err != nil {
		return nil, errors.New("invalid value for lessonId")
	}
	afterLessonID, err := mytype.ParseOID(args.Input.AfterLessonID)
	if err != nil {
		return nil, errors.New("invalid value for afterLessonId")
	}

	_, err = r.Repos.CourseLesson().Move(ctx, courseLesson, afterLessonID.String)
	if err != nil {
		return nil, err
	}

	return &moveCourseLessonPayloadResolver{
		Conf:     r.Conf,
		CourseID: &courseLesson.CourseID,
		LessonID: &courseLesson.LessonID,
		Repos:    r.Repos,
	}, nil
}

type PublishLessonDraftInput struct {
	LessonID string
}

func (r *RootResolver) PublishLessonDraft(
	ctx context.Context,
	args struct{ Input PublishLessonDraftInput },
) (*lessonResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	currentLessonPermit, err := r.Repos.Lesson().Get(ctx, args.Input.LessonID)
	if err != nil {
		return nil, errors.New("lesson not found")
	}
	draft, err := currentLessonPermit.Draft()
	if err != nil {
		return nil, err
	}
	studyID, err := currentLessonPermit.StudyID()
	if err != nil {
		return nil, err
	}
	userID, err := currentLessonPermit.UserID()
	if err != nil {
		return nil, err
	}

	lesson := &data.Lesson{}
	if err := lesson.ID.Set(args.Input.LessonID); err != nil {
		return nil, errors.New("Invalid lessonId")
	}

	if err := lesson.Body.Set(draft); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}

	err = r.Repos.ParseLessonBodyForEvents(
		ctx,
		&lesson.Body,
		&lesson.ID,
		studyID,
		userID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	body, err, updated := r.Repos.ReplaceMarkdownRefsWithLinks(
		ctx,
		lesson.Body,
		studyID.String,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if updated {
		if err := lesson.Body.Set(body); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	if err := lesson.PublishedAt.Set(time.Now()); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}

	lessonPermit, err := r.Repos.Lesson().Update(ctx, lesson)
	if err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &lessonResolver{
		Lesson: lessonPermit,
		Conf:   r.Conf,
		Repos:  r.Repos,
	}, nil
}

type PublishLessonCommentDraftInput struct {
	LessonCommentID string
}

func (r *RootResolver) PublishLessonCommentDraft(
	ctx context.Context,
	args struct {
		Input PublishLessonCommentDraftInput
	},
) (*lessonCommentResolver, error) {
	currentLessonCommentPermit, err := r.Repos.LessonComment().Get(ctx, args.Input.LessonCommentID)
	if err != nil {
		return nil, errors.New("lesson comment not found")
	}
	draft, err := currentLessonCommentPermit.Draft()
	if err != nil {
		return nil, err
	}
	studyID, err := currentLessonCommentPermit.StudyID()
	if err != nil {
		return nil, err
	}
	userID, err := currentLessonCommentPermit.UserID()
	if err != nil {
		return nil, err
	}

	lessonComment := &data.LessonComment{}
	if err := lessonComment.ID.Set(args.Input.LessonCommentID); err != nil {
		return nil, errors.New("Invalid lessonCommentId")
	}

	if err := lessonComment.Body.Set(draft); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}

	err = r.Repos.ParseLessonBodyForEvents(
		ctx,
		&lessonComment.Body,
		&lessonComment.LessonID,
		studyID,
		userID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	body, err, updated := r.Repos.ReplaceMarkdownRefsWithLinks(
		ctx,
		lessonComment.Body,
		studyID.String,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if updated {
		if err := lessonComment.Body.Set(body); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	if err := lessonComment.PublishedAt.Set(time.Now()); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}

	lessonCommentPermit, err := r.Repos.LessonComment().Update(ctx, lessonComment)
	if err != nil {
		return nil, err
	}

	return &lessonCommentResolver{
		Conf:          r.Conf,
		LessonComment: lessonCommentPermit,
		Repos:         r.Repos,
	}, nil
}

type RemoveCourseLessonInput struct {
	CourseID string
	LessonID string
}

func (r *RootResolver) RemoveCourseLesson(
	ctx context.Context,
	args struct{ Input RemoveCourseLessonInput },
) (*removeCourseLessonPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		err := errors.New("viewer not found")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	courseLesson := &data.CourseLesson{}
	if err := courseLesson.CourseID.Set(args.Input.CourseID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("invalid value for courseId")
	}
	if err := courseLesson.LessonID.Set(args.Input.LessonID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("invalid value for lessonId")
	}

	if err = r.Repos.CourseLesson().Disconnect(ctx, courseLesson); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	lesson, err := r.Repos.Lesson().Pull(ctx, args.Input.LessonID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("lesson not found")
	}
	studyID, err := lesson.StudyID()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}

	eventPayload, err := data.NewLessonRemovedFromCoursePayload(
		&courseLesson.LessonID,
		&courseLesson.CourseID,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	event, err := data.NewLessonEvent(eventPayload, studyID, &viewer.ID, true)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.Repos.Event().Create(ctx, event); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	return &removeCourseLessonPayloadResolver{
		Conf:     r.Conf,
		CourseID: &courseLesson.CourseID,
		LessonID: &courseLesson.LessonID,
		Repos:    r.Repos,
	}, nil
}

type RemoveLabelInput struct {
	LabelID     string
	LabelableID string
}

func (r *RootResolver) RemoveLabel(
	ctx context.Context,
	args struct{ Input RemoveLabelInput },
) (*removeLabelPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	labeled := &data.Labeled{}
	if err := labeled.LabelID.Set(args.Input.LabelID); err != nil {
		return nil, errors.New("invalid labeled label_id")
	}
	if err := labeled.LabelableID.Set(args.Input.LabelableID); err != nil {
		return nil, errors.New("invalid labeled labelable_id")
	}

	if err := r.Repos.Labeled().Disconnect(ctx, labeled); err != nil {
		return nil, err
	}

	if labeled.LabelableID.Type == mytype.LessonNodeType.String() {
		lesson, err := r.Repos.Lesson().Pull(ctx, args.Input.LabelableID)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, errors.New("lesson not found")
		}
		isPublished, err := lesson.IsPublished()
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, myerr.SomethingWentWrongError
		}
		studyID, err := lesson.StudyID()
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, myerr.SomethingWentWrongError
		}

		eventPayload, err := data.NewLessonUnlabeledPayload(
			&labeled.LabelableID,
			&labeled.LabelID,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		isPublic := isPublished
		event, err := data.NewLessonEvent(eventPayload, studyID, &viewer.ID, isPublic)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		if _, err := r.Repos.Event().Create(ctx, event); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &removeLabelPayloadResolver{
		Conf:        r.Conf,
		LabelID:     &labeled.LabelID,
		LabelableID: &labeled.LabelableID,
		Repos:       r.Repos,
	}, nil
}

type RequestEmailVerificationInput struct {
	Email string
}

func (r *RootResolver) RequestEmailVerification(
	ctx context.Context,
	args struct{ Input RequestEmailVerificationInput },
) (bool, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return false, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}

	email, err := data.GetEmailByValue(tx, args.Input.Email)
	if err != nil {
		if err == data.ErrNotFound {
			return false, errors.New("email not found")
		}
		return false, err
	}

	user, err := data.GetUser(tx, email.UserID.String)
	if err != nil {
		return false, err
	}

	if email.VerifiedAt.Status != pgtype.Null {
		return false, errors.New("email already verified")
	}

	evt := &data.EVT{}
	if err := evt.EmailID.Set(&email.ID); err != nil {
		return false, myerr.UnexpectedError{"failed to set evt email_id"}
	}
	if err := evt.UserID.Set(&email.UserID); err != nil {
		return false, myerr.UnexpectedError{"failed to set evt user_id"}
	}

	_, err = data.CreateEVT(tx, evt)
	if err != nil {
		return false, err
	}

	sendMailInput := &service.SendEmailVerificationMailInput{
		EmailID:   email.ID.Short,
		To:        email.Value.String,
		UserLogin: user.Login.String,
		Token:     evt.Token.String,
	}
	err = r.Svcs.Mail.SendEmailVerificationMail(sendMailInput)
	if err != nil {
		return false, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

type RequestPasswordResetInput struct {
	Email string
}

func (r *RootResolver) RequestPasswordReset(
	ctx context.Context,
	args struct{ Input RequestPasswordResetInput },
) (*prtResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	email, err := data.GetEmailByValue(tx, args.Input.Email)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, errors.New("email not found")
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}
	user, err := data.GetUser(tx, email.UserID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, errors.New("no user with that email was found")
	}

	ctx = myctx.NewUserContext(ctx, user)

	requestIp, ok := myctx.RequesterIpFromContext(ctx)
	if !ok {
		err := errors.New("requester ip not found")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	prt := &data.PRT{}
	if err := prt.EmailID.Set(&email.ID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}
	if err := prt.UserID.Set(&email.UserID); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}
	if err := prt.RequestIP.Set(requestIp); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, myerr.SomethingWentWrongError
	}

	prtPermit, err := r.Repos.PRT().Create(ctx, prt)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	token, err := prtPermit.Token()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	resolver := &prtResolver{
		Conf:  r.Conf,
		PRT:   prtPermit,
		Repos: r.Repos,
	}

	sendMailInput := &service.SendPasswordResetInput{
		To:        args.Input.Email,
		UserLogin: user.Login.String,
		Token:     token,
	}
	err = r.Svcs.Mail.SendPasswordResetMail(sendMailInput)
	if err != nil {
		return resolver, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return resolver, err
		}
	}

	return resolver, nil
}

type ResetLessonDraftInput struct {
	LessonID string
}

func (r *RootResolver) ResetLessonDraft(
	ctx context.Context,
	args struct{ Input ResetLessonDraftInput },
) (*lessonResolver, error) {
	currentLessonPermit, err := r.Repos.Lesson().Get(ctx, args.Input.LessonID)
	if err != nil {
		return nil, errors.New("lesson not found")
	}
	body, err := currentLessonPermit.Body()
	if err != nil {
		return nil, err
	}
	studyID, err := currentLessonPermit.StudyID()
	if err != nil {
		return nil, err
	}

	lesson := &data.Lesson{}
	if err := lesson.ID.Set(args.Input.LessonID); err != nil {
		return nil, errors.New("Invalid lessonId")
	}

	draft, err, updated := r.Repos.ReplaceMarkdownLinksWithRefs(
		ctx,
		body.String,
		studyID.String,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if updated {
		if err := lesson.Draft.Set(draft); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	lessonPermit, err := r.Repos.Lesson().Update(ctx, lesson)
	if err != nil {
		return nil, err
	}

	return &lessonResolver{
		Conf:   r.Conf,
		Lesson: lessonPermit,
		Repos:  r.Repos,
	}, nil
}

type ResetLessonCommentDraftInput struct {
	LessonCommentID string
}

func (r *RootResolver) ResetLessonCommentDraft(
	ctx context.Context,
	args struct {
		Input ResetLessonCommentDraftInput
	},
) (*lessonCommentResolver, error) {
	currentLessonCommentPermit, err := r.Repos.LessonComment().Get(
		ctx,
		args.Input.LessonCommentID,
	)
	if err != nil {
		return nil, errors.New("lesson comment not found")
	}
	body, err := currentLessonCommentPermit.Body()
	if err != nil {
		return nil, err
	}
	studyID, err := currentLessonCommentPermit.StudyID()
	if err != nil {
		return nil, err
	}

	lessonComment := &data.LessonComment{}
	if err := lessonComment.ID.Set(args.Input.LessonCommentID); err != nil {
		return nil, errors.New("Invalid lessonCommentId")
	}

	draft, err, updated := r.Repos.ReplaceMarkdownLinksWithRefs(
		ctx,
		body.String,
		studyID.String,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if updated {
		if err := lessonComment.Draft.Set(draft); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	lessonCommentPermit, err := r.Repos.LessonComment().Update(ctx, lessonComment)
	if err != nil {
		return nil, err
	}

	return &lessonCommentResolver{
		Conf:          r.Conf,
		LessonComment: lessonCommentPermit,
		Repos:         r.Repos,
	}, nil
}

type ResetPasswordInput struct {
	Email    string
	Token    string
	Password string
}

func (r *RootResolver) ResetPassword(
	ctx context.Context,
	args struct{ Input ResetPasswordInput },
) (bool, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return false, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	user, err := data.GetUserCredentialsByEmail(tx, args.Input.Email)
	if err != nil {
		return false, err
	}

	ctx = myctx.NewUserContext(ctx, user)

	prtPermit, err := r.Repos.PRT().Get(ctx, user.ID.String, args.Input.Token)
	if err != nil {
		return false, err
	}
	prt := prtPermit.Get()

	if prt.ExpiresAt.Time.Before(time.Now()) {
		return false, errors.New("token has expired")
	}

	if prt.EndedAt.Status == pgtype.Present {
		return false, errors.New("token has already ended")
	}

	if err = user.Password.Set(args.Input.Password); err != nil {
		mylog.Log.WithError(err).Error("failed to set password")
		return false, err
	}
	if err := user.Password.CheckStrength(mytype.Weak); err != nil {
		mylog.Log.WithError(err).Error("password failed strength check")
		return false, err
	}

	if _, err := r.Repos.User().UpdateAccount(ctx, user); err != nil {
		return false, myerr.UnexpectedError{"failed to update user"}
	}

	endIp, ok := myctx.RequesterIpFromContext(ctx)
	if !ok {
		return false, errors.New("requester ip not found")
	}

	if err := prt.UserID.Set(&user.ID); err != nil {
		return false, myerr.UnexpectedError{"failed to set prt user_id"}
	}
	if err := prt.Token.Set(args.Input.Token); err != nil {
		return false, myerr.UnexpectedError{"failed to set prt token"}
	}
	if err := prt.EndIP.Set(endIp); err != nil {
		return false, myerr.UnexpectedError{"failed to set prt end_ip"}
	}
	if err := prt.EndedAt.Set(time.Now()); err != nil {
		return false, myerr.UnexpectedError{"failed to set prt ended_at"}
	}

	if _, err := r.Repos.PRT().Update(ctx, prt); err != nil {
		return false, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

type TakeAppleInput struct {
	AppleableID string
}

func (r *RootResolver) TakeApple(
	ctx context.Context,
	args struct{ Input TakeAppleInput },
) (*appleableResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	appled := &data.Appled{}
	if err := appled.AppleableID.Set(args.Input.AppleableID); err != nil {
		return nil, errors.New("invalid appleable id")
	}
	if err := appled.UserID.Set(&viewer.ID); err != nil {
		return nil, errors.New("invalid appleable user_id")
	}
	err = r.Repos.Appled().Disconnect(ctx, appled)
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetAppleable(ctx, &appled.AppleableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	appleable, ok := resolver.(appleable)
	if !ok {
		return nil, errors.New("cannot convert resolver to appleable")
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &appleableResolver{appleable}, nil
}

type UpdateEmailInput struct {
	EmailID string
	Type    *string
}

func (r *RootResolver) UpdateEmail(
	ctx context.Context,
	args struct{ Input UpdateEmailInput },
) (*emailResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	emailPermit, err := r.Repos.Email().Get(ctx, args.Input.EmailID)
	if err != nil {
		return nil, err
	}
	email := emailPermit.Get()

	if email.VerifiedAt.Status == pgtype.Null {
		return nil, errors.New("cannot update unverified email")
	}

	emailType, err := mytype.ParseEmailType(*args.Input.Type)
	if err != nil {
		return nil, errors.New("invalid email type")
	}
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	if args.Input.Type != nil {
		if emailType.V == mytype.PrimaryEmail {
			filters := &data.EmailFilterOptions{
				Types: &[]string{emailType.String()},
			}
			emails, err := r.Repos.Email().GetByUser(ctx, viewer.ID.String, nil, filters)
			if err != nil || len(emails) == 0 {
				return nil, err
			}
			e := emails[0].Get()
			if err := e.Type.Set(mytype.ExtraEmail); err != nil {
				return nil, myerr.UnexpectedError{"failed to set email type"}
			}
			_, err = r.Repos.Email().Update(ctx, e)
			if err != nil {
				return nil, err
			}
		}
		if emailType.V == mytype.BackupEmail {
			filters := &data.EmailFilterOptions{
				Types: &[]string{emailType.String()},
			}
			emails, err := r.Repos.Email().GetByUser(ctx, viewer.ID.String, nil, filters)
			if err != nil {
				return nil, err
			}
			if len(emails) > 0 {
				e := emails[0].Get()
				if err := e.Type.Set(mytype.ExtraEmail); err != nil {
					return nil, myerr.UnexpectedError{"failed to set email type"}
				}
				_, err = r.Repos.Email().Update(ctx, e)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	emailUpdate := &data.Email{}
	if err := emailUpdate.ID.Set(&email.ID); err != nil {
		return nil, myerr.UnexpectedError{"failed to set email id"}
	}
	if err := emailUpdate.UserID.Set(&viewer.ID); err != nil {
		return nil, myerr.UnexpectedError{"failed to set email id"}
	}
	if args.Input.Type != nil {
		if err := emailUpdate.Type.Set(emailType); err != nil {
			return nil, myerr.UnexpectedError{"failed to set email type"}
		}
	}

	emailPermit, err = r.Repos.Email().Update(ctx, emailUpdate)
	if err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &emailResolver{
		Conf:  r.Conf,
		Email: emailPermit,
		Repos: r.Repos,
	}, nil
}

type UpdateEnrollmentInput struct {
	EnrollableID string
	Status       string
}

func (r *RootResolver) UpdateEnrollment(
	ctx context.Context,
	args struct{ Input UpdateEnrollmentInput },
) (*enrollableResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	enrolled := &data.Enrolled{}
	if err := enrolled.EnrollableID.Set(args.Input.EnrollableID); err != nil {
		return nil, errors.New("invalid enrollable id")
	}
	if err := enrolled.UserID.Set(&viewer.ID); err != nil {
		return nil, errors.New("invalid enrollable user_id")
	}
	if err := enrolled.Status.Set(args.Input.Status); err != nil {
		return nil, errors.New("invalid enrolled status")
	}
	if _, err := r.Repos.Enrolled().Pull(ctx, enrolled); err != nil {
		if err != data.ErrNotFound {
			return nil, err
		}
		if err := enrolled.ReasonName.Set(data.ManualReason); err != nil {
			return nil, errors.New("invalid enrolled status")
		}
		_, err := r.Repos.Enrolled().Connect(ctx, enrolled)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := r.Repos.Enrolled().Update(ctx, enrolled)
		if err != nil {
			return nil, err
		}
	}
	permit, err := r.Repos.GetEnrollable(ctx, &enrolled.EnrollableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	enrollable, ok := resolver.(enrollable)
	if !ok {
		return nil, errors.New("cannot convert resolver to enrollable")
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &enrollableResolver{enrollable}, nil
}

type UpdateLabelInput struct {
	Color       *string
	Description *string
	LabelID     string
}

func (r *RootResolver) UpdateLabel(
	ctx context.Context,
	args struct{ Input UpdateLabelInput },
) (*labelResolver, error) {
	label := &data.Label{}
	if err := label.ID.Set(args.Input.LabelID); err != nil {
		return nil, errors.New("Invalid labelId")
	}

	if args.Input.Color != nil {
		if err := label.Color.Set(args.Input.Color); err != nil {
			return nil, errors.New("Invalid color")
		}
	}
	if args.Input.Description != nil {
		if err := label.Description.Set(args.Input.Description); err != nil {
			return nil, errors.New("Invalid description")
		}
	}

	labelPermit, err := r.Repos.Label().Update(ctx, label)
	if err != nil {
		return nil, err
	}
	return &labelResolver{
		Conf:  r.Conf,
		Label: labelPermit,
		Repos: r.Repos,
	}, nil
}

type UpdateLessonInput struct {
	Draft    *string
	LessonID string
	Title    *string
}

func (r *RootResolver) UpdateLesson(
	ctx context.Context,
	args struct{ Input UpdateLessonInput },
) (*lessonResolver, error) {
	lesson := &data.Lesson{}
	if err := lesson.ID.Set(args.Input.LessonID); err != nil {
		mylog.Log.WithError(err).Error("failed to set lesson id")
		return nil, errors.New("Invalid lessonId")
	}
	if args.Input.Draft != nil {
		if err := lesson.Draft.Set(args.Input.Draft); err != nil {
			mylog.Log.WithError(err).Error("failed to set lesson draft")
			return nil, errors.New("Invalid draft")
		}
	}
	if args.Input.Title != nil {
		if err := lesson.Title.Set(args.Input.Title); err != nil {
			mylog.Log.WithError(err).Error("failed to set lesson title")
			return nil, errors.New("Invalid title")
		}
	}

	lessonPermit, err := r.Repos.Lesson().Update(ctx, lesson)
	if err != nil {
		return nil, err
	}

	return &lessonResolver{
		Conf:   r.Conf,
		Lesson: lessonPermit,
		Repos:  r.Repos,
	}, nil
}

type UpdateLessonCommentInput struct {
	Draft           *string
	LessonCommentID string
}

func (r *RootResolver) UpdateLessonComment(
	ctx context.Context,
	args struct{ Input UpdateLessonCommentInput },
) (*lessonCommentResolver, error) {
	lessonComment := &data.LessonComment{}
	if err := lessonComment.ID.Set(args.Input.LessonCommentID); err != nil {
		mylog.Log.WithError(err).Error("failed to set lesson comment id")
		return nil, errors.New("Invalid lessonCommentId")
	}
	if args.Input.Draft != nil {
		if err := lessonComment.Draft.Set(args.Input.Draft); err != nil {
			mylog.Log.WithError(err).Error("failed to set lesson comment draft")
			return nil, errors.New("Invalid draft")
		}
	}

	lessonCommentPermit, err := r.Repos.LessonComment().Update(ctx, lessonComment)
	if err != nil {
		return nil, err
	}

	return &lessonCommentResolver{
		Conf:          r.Conf,
		LessonComment: lessonCommentPermit,
		Repos:         r.Repos,
	}, nil
}

type UpdateCourseInput struct {
	Description *string
	Name        *string
	CourseID    string
}

func (r *RootResolver) UpdateCourse(
	ctx context.Context,
	args struct{ Input UpdateCourseInput },
) (*courseResolver, error) {
	course := &data.Course{}
	if err := course.ID.Set(args.Input.CourseID); err != nil {
		return nil, errors.New("invalid course id")
	}

	if args.Input.Description != nil {
		if err := course.Description.Set(args.Input.Description); err != nil {
			return nil, errors.New("invalid course description")
		}
	}
	if args.Input.Name != nil {
		if err := course.Name.Set(args.Input.Name); err != nil {
			return nil, errors.New("invalid course name")
		}
	}

	coursePermit, err := r.Repos.Course().Update(ctx, course)
	if err != nil {
		return nil, err
	}
	return &courseResolver{
		Conf:   r.Conf,
		Course: coursePermit,
		Repos:  r.Repos,
	}, nil
}

type UpdateStudyInput struct {
	Description *string
	Name        *string
	StudyID     string
}

func (r *RootResolver) UpdateStudy(
	ctx context.Context,
	args struct{ Input UpdateStudyInput },
) (*studyResolver, error) {
	study := &data.Study{}
	if err := study.ID.Set(args.Input.StudyID); err != nil {
		return nil, myerr.UnexpectedError{"failed to set study id"}
	}

	if args.Input.Description != nil {
		if err := study.Description.Set(args.Input.Description); err != nil {
			return nil, myerr.UnexpectedError{"failed to set study description"}
		}
	}
	if args.Input.Name != nil {
		if err := study.Name.Set(args.Input.Name); err != nil {
			return nil, myerr.UnexpectedError{"failed to set study name"}
		}
	}

	studyPermit, err := r.Repos.Study().Update(ctx, study)
	if err != nil {
		return nil, err
	}
	return &studyResolver{
		Conf:  r.Conf,
		Study: studyPermit,
		Repos: r.Repos,
	}, nil
}

type UpdateTopicInput struct {
	Description string
	TopicID     string
}

func (r *RootResolver) UpdateTopic(
	ctx context.Context,
	args struct{ Input UpdateTopicInput },
) (*topicResolver, error) {
	topic := &data.Topic{}
	if err := topic.ID.Set(args.Input.TopicID); err != nil {
		return nil, myerr.UnexpectedError{"failed to set topic id"}
	}
	if err := topic.Description.Set(args.Input.Description); err != nil {
		return nil, myerr.UnexpectedError{"failed to set topic description"}
	}

	topicPermit, err := r.Repos.Topic().Update(ctx, topic)
	if err != nil {
		return nil, err
	}
	return &topicResolver{
		Conf:  r.Conf,
		Topic: topicPermit,
		Repos: r.Repos,
	}, nil
}

type UpdateTopicsInput struct {
	Description *string
	TopicableID string
	TopicNames  []string
}

func (r *RootResolver) UpdateTopics(
	ctx context.Context,
	args struct{ Input UpdateTopicsInput },
) (*updateTopicsPayloadResolver, error) {
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	topicableID, err := mytype.ParseOID(args.Input.TopicableID)
	if err != nil {
		return nil, err
	}
	resolver := &updateTopicsPayloadResolver{
		Conf:        r.Conf,
		Repos:       r.Repos,
		TopicableID: topicableID,
	}
	newTopics := make(map[string]struct{})
	oldTopics := make(map[string]struct{})
	// remove empty strings from topic names
	topicNames := make([]string, 0, len(args.Input.TopicableID))
	for _, t := range args.Input.TopicNames {
		if strings.TrimSpace(t) != "" {
			topicNames = append(topicNames, t)
		}
	}
	invalidTopicNames := validateTopicNames(topicNames)
	if len(invalidTopicNames) > 0 {
		resolver.InvalidNames = invalidTopicNames
		return resolver, nil
	}
	topicPermits, err := r.Repos.Topic().GetByTopicable(
		ctx,
		args.Input.TopicableID,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	topics := make([]*data.Topic, len(topicPermits))
	for i, tp := range topicPermits {
		topics[i] = tp.Get()
		oldTopics[topics[i].Name.String] = struct{}{}
	}
	for _, name := range topicNames {
		newTopics[name] = struct{}{}
		if _, prs := oldTopics[name]; !prs {
			t := &data.Topic{}
			t.Name.Set(name)
			if err := t.Name.Set(name); err != nil {
				return nil, errors.New("invalid topic name")
			}
			topic, err := r.Repos.Topic().Create(ctx, t)
			if err != nil {
				return nil, err
			}
			topicID, err := topic.ID()
			if err != nil {
				return nil, err
			}
			topiced := &data.Topiced{}
			if err := topiced.TopicID.Set(topicID); err != nil {
				return nil, errors.New("invalid topic id")
			}
			if err := topiced.TopicableID.Set(args.Input.TopicableID); err != nil {
				return nil, errors.New("invalid topicable id")
			}
			_, err = r.Repos.Topiced().Connect(ctx, topiced)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, t := range topics {
		name := t.Name.String
		if _, prs := newTopics[name]; !prs {
			topiced := &data.Topiced{}
			if err := topiced.TopicID.Set(&t.ID); err != nil {
				return nil, errors.New("invalid topic id")
			}
			if err := topiced.TopicableID.Set(topicableID); err != nil {
				return nil, errors.New("invalid topicable id")
			}
			err := r.Repos.Topiced().Disconnect(ctx, topiced)
			if err != nil {
				return nil, err
			}
		}
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return resolver, nil
}

var validTopicName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9|-]+[a-zA-Z0-9]$`)

func validateTopicNames(topicNames []string) (invalidTopicNames []string) {
	invalidTopicNames = make([]string, 0, len(topicNames))
	for _, name := range topicNames {
		if ok := validTopicName.MatchString(name); !ok {
			invalidTopicNames = append(invalidTopicNames, name)
		}
	}
	return
}

type UpdateUserAssetInput struct {
	Description *string
	Name        *string
	UserAssetID string
}

func (r *RootResolver) UpdateUserAsset(
	ctx context.Context,
	args struct{ Input UpdateUserAssetInput },
) (*userAssetResolver, error) {
	userAsset := &data.UserAsset{}
	if err := userAsset.ID.Set(args.Input.UserAssetID); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user_asset id"}
	}

	if args.Input.Description != nil {
		if err := userAsset.Description.Set(args.Input.Description); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user_asset description"}
		}
	}
	if args.Input.Name != nil {
		if err := userAsset.Name.Set(args.Input.Name); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user_asset name"}
		}
	}

	userAssetPermit, err := r.Repos.UserAsset().Update(ctx, userAsset)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{
		Conf:      r.Conf,
		Repos:     r.Repos,
		UserAsset: userAssetPermit,
	}, nil
}

type UpdateViewerAccountInput struct {
	Login       *string
	NewPassword *string
	OldPassword *string
}

func (r *RootResolver) UpdateViewerAccount(
	ctx context.Context,
	args struct{ Input UpdateViewerAccountInput },
) (*userResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	user := &data.User{}
	if err := user.ID.Set(&viewer.ID); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user id"}
	}
	if args.Input.NewPassword != nil && args.Input.OldPassword != nil {
		if err := viewer.Password.CompareToPassword(*args.Input.OldPassword); err != nil {
			return nil, errors.New("incorrect password")
		}
		if err := user.Password.Set(args.Input.NewPassword); err != nil {
			mylog.Log.WithError(err).Error("failed to set password")
			return nil, err
		}
		if err := user.Password.CheckStrength(mytype.Weak); err != nil {
			mylog.Log.WithError(err).Error("password failed strength check")
			return nil, err
		}
	}

	if args.Input.Login != nil {
		if err := user.Login.Set(args.Input.Login); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user login"}
		}
	}
	ok, err := user.Login.IsBlacklisted()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	} else if ok {
		err := errors.New("username unavailable")
		mylog.Log.WithError(err).Error(util.Trace("failed blacklist check"))
		return nil, err
	}

	userPermit, err := r.Repos.User().UpdateAccount(ctx, user)
	if err != nil {
		return nil, err
	}
	return &userResolver{
		Conf:  r.Conf,
		Repos: r.Repos,
		User:  userPermit,
	}, nil
}

type UpdateViewerProfileInput struct {
	Bio     *string
	EmailID *string
	Name    *string
}

func (r *RootResolver) UpdateViewerProfile(
	ctx context.Context,
	args struct{ Input UpdateViewerProfileInput },
) (*userResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	user := &data.User{}
	if err := user.ID.Set(&viewer.ID); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user id"}
	}

	if args.Input.Bio != nil {
		if err := user.Bio.Set(args.Input.Bio); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user bio"}
		}
	}
	if args.Input.EmailID != nil {
		emailID := args.Input.EmailID
		// If the email ID is empty, then set pointer to nil. This will ensure a
		// NULL db value.
		if *emailID == "" {
			emailID = nil
		}
		if err := user.ProfileEmailID.Set(args.Input.EmailID); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user profile_email_id"}
		}
	}
	if args.Input.Name != nil {
		if err := user.Name.Set(args.Input.Name); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user name"}
		}
	}

	userPermit, err := r.Repos.User().UpdateProfile(ctx, user)
	if err != nil {
		return nil, err
	}
	return &userResolver{
		Conf:  r.Conf,
		User:  userPermit,
		Repos: r.Repos,
	}, nil
}

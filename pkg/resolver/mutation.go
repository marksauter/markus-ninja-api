package resolver

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/badoux/checkmail"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

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
		return nil, errors.New("invalid email value")
	}
	email.UserId.Set(&viewer.Id)
	if err := email.UserId.Set(&viewer.Id); err != nil {
		return nil, myerr.UnexpectedError{"failed to set email user_id"}
	}

	emailPermit, err := r.Repos.Email().Create(ctx, email)
	if err != nil {
		return nil, err
	}

	evt := &data.EVT{}
	if err := evt.EmailId.Set(&email.Id); err != nil {
		return nil, myerr.UnexpectedError{"failed to set evt email_id"}
	}
	if err := evt.UserId.Set(&viewer.Id); err != nil {
		return nil, myerr.UnexpectedError{"failed to set evt user_id"}
	}

	evtPermit, err := r.Repos.EVT().Create(ctx, evt)
	if err != nil {
		return nil, err
	}

	resolver := &addEmailPayloadResolver{
		Email: emailPermit,
		EVT:   evtPermit,
		Repos: r.Repos,
	}
	sendMailInput := &service.SendEmailVerificationMailInput{
		EmailId:   email.Id.Short,
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
	LabelId     string
	LabelableId string
}

func (r *RootResolver) AddLabel(
	ctx context.Context,
	args struct{ Input AddLabelInput },
) (*addLabelPayloadResolver, error) {
	labeled := &data.Labeled{}
	if err := labeled.LabelId.Set(args.Input.LabelId); err != nil {
		return nil, errors.New("invalid labeled label_id")
	}
	if err := labeled.LabelableId.Set(args.Input.LabelableId); err != nil {
		return nil, errors.New("invalid labeled labelable_id")
	}

	_, err := r.Repos.Labeled().Connect(ctx, labeled)
	if err != nil {
		return nil, err
	}

	return &addLabelPayloadResolver{
		LabelId:     &labeled.LabelId,
		LabelableId: &labeled.LabelableId,
		Repos:       r.Repos,
	}, nil
}

type AddLessonCommentInput struct {
	Body     string
	LessonId string
}

func (r *RootResolver) AddLessonComment(
	ctx context.Context,
	args struct{ Input AddLessonCommentInput },
) (*addLessonCommentPayloadResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	lessonComment := &data.LessonComment{}
	if err := lessonComment.Body.Set(args.Input.Body); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson_comment body"}
	}
	if err := lessonComment.LessonId.Set(args.Input.LessonId); err != nil {
		return nil, errors.New("invalid lesson_comment lesson_id")
	}
	if err := lessonComment.UserId.Set(&viewer.Id); err != nil {
		return nil, errors.New("invalid lesson_comment user_id")
	}

	lessonCommentPermit, err := r.Repos.LessonComment().Create(ctx, lessonComment)
	if err != nil {
		return nil, err
	}

	return &addLessonCommentPayloadResolver{
		LessonComment: lessonCommentPermit,
		Repos:         r.Repos,
	}, nil
}

type CreateLabelInput struct {
	Color       string
	Description *string
	Name        string
	StudyId     string
}

func (r *RootResolver) CreateLabel(
	ctx context.Context,
	args struct{ Input CreateLabelInput },
) (*labelResolver, error) {
	label := &data.Label{}
	if err := label.Color.Set(args.Input.Color); err != nil {
		return nil, err
	}
	if err := label.Description.Set(args.Input.Description); err != nil {
		return nil, err
	}
	if err := label.Name.Set(args.Input.Name); err != nil {
		return nil, err
	}
	if err := label.StudyId.Set(args.Input.StudyId); err != nil {
		return nil, err
	}
	labelPermit, err := r.Repos.Label().Create(ctx, label)
	if err != nil {
		return nil, err
	}

	return &labelResolver{Label: labelPermit, Repos: r.Repos}, nil
}

type CreateLessonInput struct {
	Body    *string
	StudyId string
	Title   string
}

func (r *RootResolver) CreateLesson(
	ctx context.Context,
	args struct{ Input CreateLessonInput },
) (*lessonResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	lesson := &data.Lesson{}
	if err := lesson.Body.Set(args.Input.Body); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson body"}
	}
	if err := lesson.StudyId.Set(args.Input.StudyId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson study_id"}
	}
	if err := lesson.Title.Set(args.Input.Title); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson title"}
	}
	if err := lesson.UserId.Set(&viewer.Id); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson user_id"}
	}

	lessonPermit, err := r.Repos.Lesson().Create(ctx, lesson)
	if err != nil {
		return nil, err
	}

	return &lessonResolver{Lesson: lessonPermit, Repos: r.Repos}, nil
}

type CreateStudyInput struct {
	Description *string
	Name        string
}

func (r *RootResolver) CreateStudy(
	ctx context.Context,
	args struct{ Input CreateStudyInput },
) (*studyResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	study := &data.Study{}
	if err := study.Description.Set(args.Input.Description); err != nil {
		return nil, myerr.UnexpectedError{"failed to set study description"}
	}
	if err := study.Name.Set(args.Input.Name); err != nil {
		return nil, myerr.UnexpectedError{"failed to set study name"}
	}
	if err := study.UserId.Set(&viewer.Id); err != nil {
		return nil, myerr.UnexpectedError{"failed to set study user_id"}
	}

	studyPermit, err := r.Repos.Study().Create(ctx, study)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: studyPermit, Repos: r.Repos}, nil
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
		return nil, myerr.UnexpectedError{"failed to set user primary_email"}
	}
	if err := user.Login.Set(args.Input.Login); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user login"}
	}
	if err := user.Password.Set(args.Input.Password); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user password"}
	}

	userPermit, err := r.Repos.User().Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: userPermit, Repos: r.Repos}, nil
}

type DeleteEmailInput struct {
	EmailId string
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

	emailPermit, err := r.Repos.Email().Get(ctx, args.Input.EmailId)
	if err != nil {
		return nil, err
	}

	email := emailPermit.Get()

	n, err := r.Repos.Email().CountByUser(
		ctx,
		email.UserId.String,
		data.EmailIsVerified,
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

	if email.Type.Type == data.PrimaryEmail {
		var newPrimaryEmail *data.Email
		emails, err := r.Repos.Email().GetByUser(
			ctx,
			&email.UserId,
			nil,
			data.EmailIsVerified,
		)
		if err != nil {
			return nil, err
		}
		n := len(emails)
		for i, email := range emails {
			e := email.Get()
			if e.Type.Type == data.BackupEmail {
				newPrimaryEmail = e
			}
			if newPrimaryEmail == nil && i == n-1 {
				newPrimaryEmail = e
			}
		}
		newPrimaryEmail.Type.Set(data.PrimaryEmail)
		if _, err := r.Repos.Email().Update(ctx, newPrimaryEmail); err != nil {
			return nil, err
		}
	}

	resolver := &deleteEmailPayloadResolver{
		EmailId: &email.Id,
		UserId:  &email.UserId,
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
	LabelId string
}

func (r *RootResolver) DeleteLabel(
	ctx context.Context,
	args struct{ Input DeleteLabelInput },
) (*deleteLabelPayloadResolver, error) {
	label := &data.Label{}
	if err := label.Id.Set(args.Input.LabelId); err != nil {
		return nil, errors.New("invalid label id")
	}

	if err := r.Repos.Label().Delete(ctx, label); err != nil {
		return nil, err
	}

	return &deleteLabelPayloadResolver{
		LabelId: &label.Id,
		StudyId: &label.StudyId,
		Repos:   r.Repos,
	}, nil
}

type DeleteLessonInput struct {
	LessonId string
}

func (r *RootResolver) DeleteLesson(
	ctx context.Context,
	args struct{ Input DeleteLessonInput },
) (*deleteLessonPayloadResolver, error) {
	lesson := &data.Lesson{}
	if err := lesson.Id.Set(args.Input.LessonId); err != nil {
		return nil, errors.New("invalid lesson id")
	}

	if err := r.Repos.Lesson().Delete(ctx, lesson); err != nil {
		return nil, err
	}

	return &deleteLessonPayloadResolver{
		LessonId: &lesson.Id,
		StudyId:  &lesson.StudyId,
		Repos:    r.Repos,
	}, nil
}

type DeleteLessonCommentInput struct {
	LessonCommentId string
}

func (r *RootResolver) DeleteLessonComment(
	ctx context.Context,
	args struct{ Input DeleteLessonCommentInput },
) (*deleteLessonCommentPayloadResolver, error) {
	lessonComment := &data.LessonComment{}
	if err := lessonComment.Id.Set(args.Input.LessonCommentId); err != nil {
		return nil, errors.New("invalid lesson_comment id")
	}

	if err := r.Repos.LessonComment().Delete(ctx, lessonComment); err != nil {
		return nil, err
	}

	return &deleteLessonCommentPayloadResolver{
		LessonCommentId: &lessonComment.Id,
		LessonId:        &lessonComment.LessonId,
		Repos:           r.Repos,
	}, nil
}

type DeleteStudyInput struct {
	StudyId string
}

func (r *RootResolver) DeleteStudy(
	ctx context.Context,
	args struct{ Input DeleteStudyInput },
) (*deleteStudyPayloadResolver, error) {
	study := &data.Study{}
	if err := study.Id.Set(args.Input.StudyId); err != nil {
		return nil, errors.New("invalid study id")
	}

	if err := r.Repos.Study().Delete(ctx, study); err != nil {
		return nil, err
	}

	return &deleteStudyPayloadResolver{
		OwnerId: &study.UserId,
		StudyId: &study.Id,
		Repos:   r.Repos,
	}, nil
}

type DeleteUserAssetInput struct {
	UserAssetId string
}

func (r *RootResolver) DeleteUserAsset(
	ctx context.Context,
	args struct{ Input DeleteUserAssetInput },
) (*deleteUserAssetPayloadResolver, error) {
	userAsset := &data.UserAsset{}
	if err := userAsset.Id.Set(args.Input.UserAssetId); err != nil {
		return nil, errors.New("invalid user_asset id")
	}

	if err := r.Repos.UserAsset().Delete(ctx, userAsset); err != nil {
		return nil, err
	}

	return &deleteUserAssetPayloadResolver{
		UserAssetId: &userAsset.Id,
		StudyId:     &userAsset.StudyId,
		Repos:       r.Repos,
	}, nil
}

type DeleteViewerAccountInput struct {
	Login    string
	Password string
}

func (r *RootResolver) DeleteViewerAccount(
	ctx context.Context,
	args struct{ Input DeleteViewerAccountInput },
) (*graphql.ID, error) {
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

	id := graphql.ID(user.Id.String)
	return &id, nil
}

type DismissInput struct {
	EnrollableId string
}

func (r *RootResolver) Dismiss(
	ctx context.Context,
	args struct{ Input DismissInput },
) (*enrollableResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	enrolled := &data.Enrolled{}
	if err := enrolled.EnrollableId.Set(args.Input.EnrollableId); err != nil {
		return nil, errors.New("invalid enrollable id")
	}
	if err := enrolled.UserId.Set(&viewer.Id); err != nil {
		return nil, errors.New("invalid enrollable user_id")
	}
	err := r.Repos.Enrolled().Disconnect(ctx, enrolled)
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetEnrollable(ctx, &enrolled.EnrollableId)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	enrollable, ok := resolver.(enrollable)
	if !ok {
		return nil, errors.New("cannot convert resolver to enrollable")
	}
	return &enrollableResolver{enrollable}, nil
}

type EnrollInput struct {
	EnrollableId string
}

func (r *RootResolver) Enroll(
	ctx context.Context,
	args struct{ Input EnrollInput },
) (*enrollableResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, errors.New("queryer not found")
	}
	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	enrolled := &data.Enrolled{}
	if err := enrolled.EnrollableId.Set(args.Input.EnrollableId); err != nil {
		return nil, errors.New("invalid enrollable id")
	}
	if err := enrolled.UserId.Set(&viewer.Id); err != nil {
		return nil, errors.New("invalid enrollable user_id")
	}
	if _, err := data.GetEnrolledByEnrollableAndUser(
		db,
		enrolled.EnrollableId.String,
		enrolled.UserId.String,
	); err != nil {
		if err != data.ErrNotFound {
			return nil, err
		}
		if err := enrolled.ReasonName.Set(data.ManualReason); err != nil {
			return nil, errors.New("invalid enrollable reason_name")
		}
		_, err := r.Repos.Enrolled().Connect(ctx, enrolled)
		if err != nil {
			return nil, err
		}
	} else {
		if err := enrolled.Ignore.Set(false); err != nil {
			return nil, errors.New("invalid enrollable ignore")
		}
		_, err := r.Repos.Enrolled().Update(ctx, enrolled)
		if err != nil {
			return nil, err
		}
	}
	permit, err := r.Repos.GetEnrollable(ctx, &enrolled.EnrollableId)
	if err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	enrollable, ok := resolver.(enrollable)
	if !ok {
		return nil, errors.New("cannot convert resolver to enrollable")
	}

	return &enrollableResolver{enrollable}, nil
}

type GiveAppleInput struct {
	AppleableId string
}

func (r *RootResolver) GiveApple(
	ctx context.Context,
	args struct{ Input GiveAppleInput },
) (*appleableResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	appled := &data.Appled{}
	if err := appled.AppleableId.Set(args.Input.AppleableId); err != nil {
		return nil, errors.New("invalid appleable id")
	}
	if err := appled.UserId.Set(&viewer.Id); err != nil {
		return nil, errors.New("invalid appleable user_id")
	}
	_, err := r.Repos.Appled().Connect(ctx, appled)
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetAppleable(ctx, &appled.AppleableId)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	appleable, ok := resolver.(appleable)
	if !ok {
		return nil, errors.New("cannot convert resolver to appleable")
	}
	return &appleableResolver{appleable}, nil
}

type IgnoreInput struct {
	EnrollableId string
}

func (r *RootResolver) Ignore(
	ctx context.Context,
	args struct{ Input IgnoreInput },
) (*enrollableResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	enrolled := &data.Enrolled{}
	if err := enrolled.EnrollableId.Set(args.Input.EnrollableId); err != nil {
		return nil, errors.New("invalid enrollable id")
	}
	if err := enrolled.Ignore.Set(true); err != nil {
		return nil, errors.New("invalid enrollable ignore")
	}
	if err := enrolled.UserId.Set(&viewer.Id); err != nil {
		return nil, errors.New("invalid enrollable user_id")
	}
	_, err := r.Repos.Enrolled().Update(ctx, enrolled)
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetEnrollable(ctx, &enrolled.EnrollableId)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	enrollable, ok := resolver.(enrollable)
	if !ok {
		return nil, errors.New("cannot convert resolver to enrollable")
	}
	return &enrollableResolver{enrollable}, nil
}

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
			return nil, errors.New("invalid credentials")
		}
	} else {
		user, err = data.GetUserCredentialsByEmail(db, args.Input.Login)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
	}

	if err := user.Password.CompareToPassword(args.Input.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	exp := time.Now().Add(time.Hour * time.Duration(24)).Unix()
	payload := myjwt.Payload{Exp: exp, Iat: time.Now().Unix(), Sub: user.Id.String}
	jwt, err := r.Svcs.Auth.SignJWT(&payload)
	if err != nil {
		return nil, err
	}

	return &loginUserPayloadResolver{
		AccessToken: jwt,
		Viewer:      user,
		Repos:       r.Repos,
	}, nil
}

type MoveLessonInput struct {
	LessonId string
	Number   *int32
}

func (r *RootResolver) MoveLesson(
	ctx context.Context,
	args struct{ Input MoveLessonInput },
) (*lessonEdgeResolver, error) {
	lesson := &data.Lesson{}
	if err := lesson.Id.Set(args.Input.LessonId); err != nil {
		return nil, errors.New("invalid lesson id")
	}
	if args.Input.Number != nil {
		if *args.Input.Number < 1 {
			return nil, errors.New("`number` must be greater than 0")
		}
		if err := lesson.Number.Set(args.Input.Number); err != nil {
			return nil, myerr.UnexpectedError{"failed to set lesson number"}
		}
	}

	lessonPermit, err := r.Repos.Lesson().Update(ctx, lesson)
	if err != nil {
		return nil, err
	}

	resolver, err := NewLessonEdgeResolver(lessonPermit, r.Repos)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

type RemoveLabelInput struct {
	LabelId     string
	LabelableId string
}

func (r *RootResolver) RemoveLabel(
	ctx context.Context,
	args struct{ Input RemoveLabelInput },
) (*removeLabelPayloadResolver, error) {
	labeled := &data.Labeled{}
	if err := labeled.LabelId.Set(args.Input.LabelId); err != nil {
		return nil, errors.New("invalid labeled label_id")
	}
	if err := labeled.LabelableId.Set(args.Input.LabelableId); err != nil {
		return nil, errors.New("invalid labeled labelable_id")
	}

	err := r.Repos.Labeled().Disconnect(ctx, labeled)
	if err != nil {
		return nil, err
	}

	return &removeLabelPayloadResolver{
		LabelId:     &labeled.LabelId,
		LabelableId: &labeled.LabelableId,
		Repos:       r.Repos,
	}, nil
}

type RequestEmailVerificationInput struct {
	Email string
}

func (r *RootResolver) RequestEmailVerification(
	ctx context.Context,
	args struct{ Input RequestEmailVerificationInput },
) (*evtResolver, error) {
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

	email, err := r.Repos.Email().GetByValue(ctx, args.Input.Email)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, errors.New("`email` not found")
		}
		return nil, err
	}
	userId, err := email.UserId()
	if err != nil {
		return nil, err
	}
	if viewer.Id.String != userId.String {
		return nil, errors.New("email already registered to another user")
	}

	isVerified, err := email.IsVerified()
	if err != nil {
		return nil, err
	}
	if isVerified {
		return nil, errors.New("email already verified")
	}

	emailId, err := email.ID()
	if err != nil {
		return nil, err
	}

	evt := &data.EVT{}
	if err := evt.EmailId.Set(emailId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set evt email_id"}
	}
	if err := evt.UserId.Set(userId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set evt user_id"}
	}

	evtPermit, err := r.Repos.EVT().Create(ctx, evt)
	if err != nil {
		return nil, err
	}

	to, err := email.Value()
	if err != nil {
		return nil, err
	}

	sendMailInput := &service.SendEmailVerificationMailInput{
		EmailId:   emailId.Short,
		To:        to,
		UserLogin: viewer.Login.String,
		Token:     evt.Token.String,
	}
	err = r.Svcs.Mail.SendEmailVerificationMail(sendMailInput)
	if err != nil {
		return nil, err
	}

	resolver := &evtResolver{EVT: evtPermit, Repos: r.Repos}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return resolver, err
		}
	}

	return resolver, nil
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
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	email, err := data.GetEmailByValue(tx, args.Input.Email)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, errors.New("`email` not found")
		}
		return nil, err
	}
	user, err := data.GetUser(tx, email.UserId.String)
	if err != nil {
		return nil, errors.New("no user with that email was found")
	}

	ctx = myctx.NewUserContext(ctx, user)

	requestIp, ok := myctx.RequesterIpFromContext(ctx)
	if !ok {
		return nil, errors.New("requester ip not found")
	}

	prt := &data.PRT{}
	if err := prt.EmailId.Set(&email.Id); err != nil {
		mylog.Log.Error(err)
		return nil, myerr.UnexpectedError{"failed to set prt email_id"}
	}
	if err := prt.UserId.Set(&email.UserId); err != nil {
		mylog.Log.Error(err)
		return nil, myerr.UnexpectedError{"failed to set prt user_id"}
	}
	if err := prt.RequestIP.Set(requestIp); err != nil {
		mylog.Log.Error(err)
		return nil, myerr.UnexpectedError{"failed to set prt request_ip"}
	}

	prtPermit, err := r.Repos.PRT().Create(ctx, prt)
	if err != nil {
		return nil, err
	}

	resolver := &prtResolver{PRT: prtPermit, Repos: r.Repos}

	sendMailInput := &service.SendPasswordResetInput{
		To:        args.Input.Email,
		UserLogin: user.Login.String,
		Token:     prt.Token.String,
	}
	err = r.Svcs.Mail.SendPasswordResetMail(sendMailInput)
	if err != nil {
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

	prtPermit, err := r.Repos.PRT().Get(ctx, user.Id.String, args.Input.Token)
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
	if err := user.Password.CheckStrength(mytype.VeryWeak); err != nil {
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

	if err := prt.UserId.Set(&user.Id); err != nil {
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
	AppleableId string
}

func (r *RootResolver) TakeApple(
	ctx context.Context,
	args struct{ Input TakeAppleInput },
) (*appleableResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	appled := &data.Appled{}
	if err := appled.AppleableId.Set(args.Input.AppleableId); err != nil {
		return nil, errors.New("invalid appleable id")
	}
	if err := appled.UserId.Set(&viewer.Id); err != nil {
		return nil, errors.New("invalid appleable user_id")
	}
	err := r.Repos.Appled().Disconnect(ctx, appled)
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetAppleable(ctx, &appled.AppleableId)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	appleable, ok := resolver.(appleable)
	if !ok {
		return nil, errors.New("cannot convert resolver to appleable")
	}
	return &appleableResolver{appleable}, nil
}

type UpdateEmailInput struct {
	EmailId string
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

	emailPermit, err := r.Repos.Email().Get(ctx, args.Input.EmailId)
	if err != nil {
		return nil, err
	}

	ok, err := emailPermit.IsVerified()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("cannot update unverified email")
	}

	if args.Input.Type != nil {
		if *args.Input.Type == data.PrimaryEmail.String() {
			viewer, ok := myctx.UserFromContext(ctx)
			if !ok {
				return nil, errors.New("viewer not found")
			}
			email, err := r.Repos.Email().GetByUserPrimary(ctx, viewer.Id.String)
			if err != nil {
				return nil, myerr.UnexpectedError{"user primary email not found"}
			}
			e := email.Get()
			if err := e.Type.Set(data.ExtraEmail); err != nil {
				return nil, myerr.UnexpectedError{"failed to set email type"}
			}
			_, err = r.Repos.Email().Update(ctx, e)
			if err != nil {
				return nil, err
			}
		}
		if *args.Input.Type == data.BackupEmail.String() {
			viewer, ok := myctx.UserFromContext(ctx)
			if !ok {
				return nil, errors.New("viewer not found")
			}
			email, err := r.Repos.Email().GetByUserBackup(ctx, viewer.Id.String)
			if err != nil && err != data.ErrNotFound {
				return nil, err
			}
			if email != nil {
				e := email.Get()
				if err := e.Type.Set(data.ExtraEmail); err != nil {
					return nil, myerr.UnexpectedError{"failed to set email type"}
				}
				_, err = r.Repos.Email().Update(ctx, e)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	email := &data.Email{}
	if err := email.Id.Set(args.Input.EmailId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set email id"}
	}
	if args.Input.Type != nil {
		if err := email.Type.Set(args.Input.Type); err != nil {
			return nil, myerr.UnexpectedError{"failed to set email type"}
		}
	}

	emailPermit, err = r.Repos.Email().Update(ctx, email)
	if err != nil {
		return nil, err
	}

	if newTx {
		err := data.CommitTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	return &emailResolver{Email: emailPermit, Repos: r.Repos}, nil
}

type UpdateLabelInput struct {
	Color       *string
	Description *string
	LabelId     string
}

func (r *RootResolver) UpdateLabel(
	ctx context.Context,
	args struct{ Input UpdateLabelInput },
) (*labelResolver, error) {
	label := &data.Label{}
	if err := label.Id.Set(args.Input.LabelId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set label id"}
	}

	if args.Input.Color != nil {
		if err := label.Color.Set(args.Input.Color); err != nil {
			return nil, myerr.UnexpectedError{"failed to set label color"}
		}
	}
	if args.Input.Description != nil {
		if err := label.Description.Set(args.Input.Description); err != nil {
			return nil, myerr.UnexpectedError{"failed to set label description"}
		}
	}

	labelPermit, err := r.Repos.Label().Update(ctx, label)
	if err != nil {
		return nil, err
	}
	return &labelResolver{Label: labelPermit, Repos: r.Repos}, nil
}

type UpdateLessonInput struct {
	Body     *string
	LessonId string
	Title    *string
}

func (r *RootResolver) UpdateLesson(
	ctx context.Context,
	args struct{ Input UpdateLessonInput },
) (*lessonResolver, error) {
	lesson := &data.Lesson{}
	if err := lesson.Id.Set(args.Input.LessonId); err != nil {
		return nil, errors.New("invalid lesson id")
	}

	if args.Input.Body != nil {
		if err := lesson.Body.Set(args.Input.Body); err != nil {
			return nil, myerr.UnexpectedError{"failed to set lesson body"}
		}
	}
	if args.Input.Title != nil {
		if err := lesson.Title.Set(args.Input.Title); err != nil {
			return nil, myerr.UnexpectedError{"failed to set lesson title"}
		}
	}

	lessonPermit, err := r.Repos.Lesson().Update(ctx, lesson)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lessonPermit, Repos: r.Repos}, nil
}

type UpdateLessonCommentInput struct {
	Body            *string
	LessonCommentId string
}

func (r *RootResolver) UpdateLessonComment(
	ctx context.Context,
	args struct{ Input UpdateLessonCommentInput },
) (*lessonCommentResolver, error) {
	lessonComment := &data.LessonComment{}
	if err := lessonComment.Id.Set(args.Input.LessonCommentId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson comment id"}
	}

	if args.Input.Body != nil {
		if err := lessonComment.Body.Set(args.Input.Body); err != nil {
			return nil, myerr.UnexpectedError{"failed to set lessonComment body"}
		}
	}

	lessonCommentPermit, err := r.Repos.LessonComment().Update(ctx, lessonComment)
	if err != nil {
		return nil, err
	}
	return &lessonCommentResolver{
		LessonComment: lessonCommentPermit,
		Repos:         r.Repos,
	}, nil
}

type UpdateStudyInput struct {
	Description *string
	Name        *string
	StudyId     string
}

func (r *RootResolver) UpdateStudy(
	ctx context.Context,
	args struct{ Input UpdateStudyInput },
) (*studyResolver, error) {
	study := &data.Study{}
	if err := study.Id.Set(args.Input.StudyId); err != nil {
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
	return &studyResolver{Study: studyPermit, Repos: r.Repos}, nil
}

type UpdateTopicInput struct {
	Description *string
	TopicId     string
}

func (r *RootResolver) UpdateTopic(
	ctx context.Context,
	args struct{ Input UpdateTopicInput },
) (*topicResolver, error) {
	topic := &data.Topic{}
	if err := topic.Id.Set(args.Input.TopicId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set topic id"}
	}
	if args.Input.Description != nil {
		if err := topic.Description.Set(args.Input.Description); err != nil {
			return nil, myerr.UnexpectedError{"failed to set topic description"}
		}
	}

	topicPermit, err := r.Repos.Topic().Update(ctx, topic)
	if err != nil {
		return nil, err
	}
	return &topicResolver{Topic: topicPermit, Repos: r.Repos}, nil
}

type UpdateTopicsInput struct {
	Description *string
	TopicableId string
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

	topicableId, err := mytype.ParseOID(args.Input.TopicableId)
	if err != nil {
		return nil, err
	}
	resolver := &updateTopicsPayloadResolver{
		TopicableId: topicableId,
		Repos:       r.Repos,
	}
	newTopics := make(map[string]struct{})
	oldTopics := make(map[string]struct{})
	invalidTopicNames := validateTopicNames(args.Input.TopicNames)
	if len(invalidTopicNames) > 0 {
		resolver.InvalidNames = invalidTopicNames
		return resolver, nil
	}
	topicPermits, err := r.Repos.Topic().GetByTopicable(
		ctx,
		args.Input.TopicableId,
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
	for _, name := range args.Input.TopicNames {
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
			topicId, err := topic.ID()
			if err != nil {
				return nil, err
			}
			topiced := &data.Topiced{}
			if err := topiced.TopicId.Set(topicId); err != nil {
				return nil, errors.New("invalid topic id")
			}
			if err := topiced.TopicableId.Set(args.Input.TopicableId); err != nil {
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
			if err := topiced.TopicId.Set(&t.Id); err != nil {
				return nil, errors.New("invalid topic id")
			}
			if err := topiced.TopicableId.Set(topicableId); err != nil {
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
	Name        *string
	UserAssetId string
}

func (r *RootResolver) UpdateUserAsset(
	ctx context.Context,
	args struct{ Input UpdateUserAssetInput },
) (*userAssetResolver, error) {
	userAsset := &data.UserAsset{}
	if err := userAsset.Id.Set(args.Input.UserAssetId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user_asset id"}
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
	return &userAssetResolver{UserAsset: userAssetPermit, Repos: r.Repos}, nil
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
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}

	var user *data.User
	var err error
	if args.Input.NewPassword != nil && args.Input.OldPassword != nil {
		user, err = data.GetUserCredentials(db, viewer.Id.String)
		if err != nil {
			return nil, err
		}
		if err := user.Password.CompareToPassword(*args.Input.OldPassword); err != nil {
			return nil, errors.New("incorrect password")
		}
		if err := user.Password.Set(args.Input.NewPassword); err != nil {
			mylog.Log.WithError(err).Error("failed to set password")
			return nil, err
		}
		if err := user.Password.CheckStrength(mytype.Strong); err != nil {
			mylog.Log.WithError(err).Error("password failed strength check")
			return nil, err
		}
	}

	if args.Input.Login != nil {
		if err := user.Login.Set(args.Input.Login); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user login"}
		}
	}

	userPermit, err := r.Repos.User().UpdateAccount(ctx, user)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: userPermit, Repos: r.Repos}, nil
}

type UpdateViewerProfileInput struct {
	Bio     *string
	EmailId *string
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
	if err := user.Id.Set(&viewer.Id); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user id"}
	}

	if args.Input.Bio != nil {
		if err := user.Bio.Set(args.Input.Bio); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user bio"}
		}
	}
	if args.Input.EmailId != nil && *args.Input.EmailId != "" {
		if err := user.ProfileEmailId.Set(args.Input.EmailId); err != nil {
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
	return &userResolver{User: userPermit, Repos: r.Repos}, nil
}

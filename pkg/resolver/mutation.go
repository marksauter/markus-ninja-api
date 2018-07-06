package resolver

import (
	"context"
	"errors"
	"regexp"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
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

	email := &data.Email{}
	email.Value.Set(args.Input.Email)
	email.UserId.Set(viewer.Id)

	emailPermit, err := r.Repos.Email().Create(email)
	if err != nil {
		return nil, err
	}

	evt := &data.EVT{}
	evt.EmailId.Set(email.Id)
	evt.UserId.Set(viewer.Id)

	evtPermit, err := r.Repos.EVT().Create(evt)
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
	labeled.LabelId.Set(args.Input.LabelId)
	labeled.LabelableId.Set(args.Input.LabelableId)

	labeledPermit, err := r.Repos.Labeled().Connect(labeled)
	if err != nil {
		return nil, err
	}

	labelPermit, err := r.Repos.Label().Get(labeled.LabelId.String)
	if err != nil {
		return nil, err
	}

	return &addLabelPayloadResolver{
		Label:   labelPermit,
		Labeled: labeledPermit,
		Repos:   r.Repos,
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

	lesson, err := r.Repos.Lesson().Get(args.Input.LessonId)
	if err != nil {
		return nil, err
	}
	studyId, err := lesson.StudyId()
	if err != nil {
		return nil, err
	}

	lessonComment := &data.LessonComment{}
	lessonComment.Body.Set(args.Input.Body)
	lessonComment.LessonId.Set(args.Input.LessonId)
	lessonComment.StudyId.Set(studyId)
	lessonComment.UserId.Set(viewer.Id)

	lessonCommentPermit, err := r.Repos.LessonComment().Create(lessonComment)
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
	labelPermit, err := r.Repos.Label().Create(label)
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
	if err := lesson.UserId.Set(viewer.Id); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson user_id"}
	}

	lessonPermit, err := r.Repos.Lesson().Create(lesson)
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
	if err := study.UserId.Set(viewer.Id.String); err != nil {
		return nil, myerr.UnexpectedError{"failed to set study user_id"}
	}

	studyPermit, err := r.Repos.Study().Create(study)
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

	if err := user.Password.Set(args.Input.Password); err != nil {
		mylog.Log.WithError(err).Error("failed to set password")
		return nil, err
	}
	if err := user.Password.CheckStrength(mytype.VeryWeak); err != nil {
		mylog.Log.WithError(err).Error("password failed strength check")
		return nil, err
	}

	if err := user.Login.Set(args.Input.Login); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user login"}
	}
	if err := user.PrimaryEmail.Set(args.Input.Email); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user primary_email"}
	}

	userPermit, err := r.Repos.User().Create(user)
	if err != nil {
		return nil, err
	}
	user = userPermit.Get()
	emailPermit, err := r.Repos.Email().GetByValue(args.Input.Email)
	primaryEmail := emailPermit.Get()

	uResolver := &userResolver{User: userPermit, Repos: r.Repos}

	if user.Login.String != "guest" {
		evt := &data.EVT{}
		evt.EmailId.Set(primaryEmail.Id)
		evt.UserId.Set(user.Id)

		newEvt, err := r.Svcs.EVT.Create(evt)
		if err != nil {
			return uResolver, err
		}

		sendMailInput := &service.SendEmailVerificationMailInput{
			EmailId:   primaryEmail.Id.String,
			To:        primaryEmail.Value.String,
			UserLogin: user.Login.String,
			Token:     newEvt.Token.String,
		}
		err = r.Svcs.Mail.SendEmailVerificationMail(sendMailInput)
		if err != nil {
			return uResolver, err
		}
	}

	return uResolver, nil
}

type DeleteEmailInput struct {
	EmailId string
}

func (r *RootResolver) DeleteEmail(
	ctx context.Context,
	args struct{ Input DeleteEmailInput },
) (*deleteEmailPayloadResolver, error) {
	emailPermit, err := r.Repos.Email().Get(args.Input.EmailId)
	if err != nil {
		return nil, err
	}

	email := emailPermit.Get()

	n, err := r.Repos.Email().CountByUser(
		email.UserId.String,
		data.EmailIsVerified,
	)
	if err != nil {
		return nil, err
	}
	if n == 1 {
		return nil, errors.New("cannot delete your only verified email")
	}

	if err := r.Repos.Email().Delete(email); err != nil {
		return nil, err
	}

	if email.Type.Type == data.PrimaryEmail {
		var newPrimaryEmail *data.Email
		emails, err := r.Repos.Email().GetByUser(
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
		if _, err := r.Repos.Email().Update(newPrimaryEmail); err != nil {
			return nil, err
		}
	}

	return &deleteEmailPayloadResolver{
		EmailId: &email.Id,
		UserId:  &email.UserId,
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
	lessonPermit, err := r.Repos.Lesson().Get(args.Input.LessonId)
	if err != nil {
		return nil, err
	}

	lesson := lessonPermit.Get()

	if err := r.Repos.Lesson().Delete(lesson); err != nil {
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
	lessonCommentPermit, err := r.Repos.LessonComment().Get(args.Input.LessonCommentId)
	if err != nil {
		return nil, err
	}

	lessonComment := lessonCommentPermit.Get()

	if err := r.Repos.LessonComment().Delete(lessonComment); err != nil {
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
	studyPermit, err := r.Repos.Study().Get(args.Input.StudyId)
	if err != nil {
		return nil, err
	}

	study := studyPermit.Get()

	if err := r.Repos.Study().Delete(study); err != nil {
		return nil, err
	}

	return &deleteStudyPayloadResolver{
		OwnerId: &study.UserId,
		StudyId: &study.Id,
		Repos:   r.Repos,
	}, nil
}

type DeleteUserInput struct {
	UserId string
}

func (r *RootResolver) DeleteUser(
	ctx context.Context,
	args struct{ Input DeleteUserInput },
) (*graphql.ID, error) {
	userPermit, err := r.Repos.User().Get(args.Input.UserId)
	if err != nil {
		return nil, err
	}

	user := userPermit.Get()

	if err := user.Id.Set(args.Input.UserId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user id"}
	}

	if err := r.Repos.User().Delete(user); err != nil {
		return nil, err
	}

	gqlID := graphql.ID(args.Input.UserId)
	return &gqlID, nil
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

	id, err := mytype.ParseOID(args.Input.EnrollableId)
	if err != nil {
		return nil, err
	}
	enrolled := &data.Enrolled{}
	enrolled.EnrollableId.Set(id)
	enrolled.UserId.Set(viewer.Id)
	err = r.Repos.Enrolled().Disconnect(enrolled)
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Study":
		study, err := r.Repos.Study().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &enrollableResolver{&studyResolver{Study: study, Repos: r.Repos}}, nil
	case "User":
		user, err := r.Repos.User().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &enrollableResolver{&userResolver{User: user, Repos: r.Repos}}, nil
	default:
		return nil, errors.New("invalid enrollable id")
	}
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

	id, err := mytype.ParseOID(args.Input.EnrollableId)
	if err != nil {
		return nil, err
	}
	enrolled := &data.Enrolled{}
	enrolled.EnrollableId.Set(id)
	enrolled.ReasonName.Set(data.ManualReason)
	enrolled.UserId.Set(viewer.Id)
	_, err = r.Repos.Enrolled().Connect(enrolled)
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Study":
		study, err := r.Repos.Study().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &enrollableResolver{&studyResolver{Study: study, Repos: r.Repos}}, nil
	case "User":
		user, err := r.Repos.User().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &enrollableResolver{&userResolver{User: user, Repos: r.Repos}}, nil
	default:
		return nil, errors.New("invalid enrollable id")
	}
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

	id, err := mytype.ParseOID(args.Input.AppleableId)
	if err != nil {
		return nil, err
	}
	appled := &data.Appled{}
	appled.AppleableId.Set(id)
	appled.UserId.Set(viewer.Id)
	_, err = r.Repos.Appled().Connect(appled)
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Study":
		study, err := r.Repos.Study().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &appleableResolver{&studyResolver{Study: study, Repos: r.Repos}}, nil
	default:
		return nil, errors.New("invalid appleable id")
	}
}

type MoveLessonInput struct {
	LessonId string
	Number   *int32
}

func (r *RootResolver) MoveLesson(
	ctx context.Context,
	args struct{ Input MoveLessonInput },
) (*lessonEdgeResolver, error) {
	number := int32(1)
	if args.Input.Number != nil {
		if *args.Input.Number < 1 {
			return nil, errors.New("`number` must be greater than 0")
		}
		number = *args.Input.Number
	}

	lessonPermit, err := r.Repos.Lesson().Get(args.Input.LessonId)
	if err != nil {
		return nil, err
	}

	lesson := lessonPermit.Get()
	if err := lesson.Number.Set(number); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson number"}
	}

	lessonPermit, err = r.Repos.Lesson().Update(lesson)
	if err != nil {
		return nil, err
	}

	resolver, err := NewLessonEdgeResolver(lessonPermit, r.Repos)
	if err != nil {
		return nil, err
	}

	return resolver, nil
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

	email, err := r.Repos.Email().GetByValue(args.Input.Email)
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

	evtPermit, err := r.Repos.EVT().Create(evt)
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

	return &evtResolver{EVT: evtPermit, Repos: r.Repos}, nil
}

type RequestPasswordResetInput struct {
	Email string
}

func (r *RootResolver) RequestPasswordReset(
	ctx context.Context,
	args struct{ Input RequestPasswordResetInput },
) (*prtResolver, error) {
	email, err := r.Svcs.Email.GetByValue(args.Input.Email)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, errors.New("`email` not found")
		}
		return nil, err
	}
	user, err := r.Svcs.User.Get(email.UserId.String)
	if err != nil {
		return nil, errors.New("no user with that email was found")
	}

	requestIp, ok := myctx.RequesterIpFromContext(ctx)
	if !ok {
		return nil, errors.New("requester ip not found")
	}

	prt := &data.PRT{}
	if err := prt.EmailId.Set(email.Id); err != nil {
		mylog.Log.Error(err)
		return nil, myerr.UnexpectedError{"failed to set prt email_id"}
	}
	if err := prt.UserId.Set(email.UserId); err != nil {
		mylog.Log.Error(err)
		return nil, myerr.UnexpectedError{"failed to set prt user_id"}
	}
	if err := prt.RequestIP.Set(requestIp); err != nil {
		mylog.Log.Error(err)
		return nil, myerr.UnexpectedError{"failed to set prt request_ip"}
	}

	prtPermit, err := r.Repos.PRT().Create(prt)
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
	user, err := r.Svcs.User.GetCredentialsByEmail(args.Input.Email)
	if err != nil {
		return false, err
	}

	prt, err := r.Svcs.PRT.Get(user.Id.String, args.Input.Token)
	if err != nil {
		return false, err
	}

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

	if _, err := r.Svcs.User.Update(user); err != nil {
		return false, myerr.UnexpectedError{"failed to update user"}
	}

	endIp, ok := myctx.RequesterIpFromContext(ctx)
	if !ok {
		return false, errors.New("requester ip not found")
	}

	if err := prt.EndIP.Set(endIp); err != nil {
		return false, myerr.UnexpectedError{"failed to set prt end_ip"}
	}
	if err := prt.EndedAt.Set(time.Now()); err != nil {
		return false, myerr.UnexpectedError{"failed to set prt ended_at"}
	}

	if _, err := r.Svcs.PRT.Update(prt); err != nil {
		return false, myerr.UnexpectedError{"failed to update prt"}
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

	id, err := mytype.ParseOID(args.Input.AppleableId)
	if err != nil {
		return nil, err
	}
	appled := &data.Appled{}
	appled.AppleableId.Set(id)
	appled.UserId.Set(viewer.Id)
	err = r.Repos.Appled().Disconnect(appled)
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Study":
		study, err := r.Repos.Study().Get(id.String)
		if err != nil {
			return nil, err
		}
		return &appleableResolver{&studyResolver{Study: study, Repos: r.Repos}}, nil
	default:
		return nil, errors.New("invalid appleable id")
	}
}

type UpdateEmailInput struct {
	EmailId string
	Type    *string
}

func (r *RootResolver) UpdateEmail(
	ctx context.Context,
	args struct{ Input UpdateEmailInput },
) (*emailResolver, error) {
	emailPermit, err := r.Repos.Email().Get(args.Input.EmailId)
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

	email := emailPermit.Get()

	if args.Input.Type != nil {
		if err := email.Type.Set(args.Input.Type); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user_email type"}
		}
	}

	emailPermit, err = r.Repos.Email().Update(email)
	if err != nil {
		return nil, err
	}
	return &emailResolver{Email: emailPermit, Repos: r.Repos}, nil
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
	lessonPermit, err := r.Repos.Lesson().Get(args.Input.LessonId)
	if err != nil {
		return nil, err
	}

	lesson := lessonPermit.Get()

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

	lessonPermit, err = r.Repos.Lesson().Update(lesson)
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
	lessonCommentPermit, err := r.Repos.LessonComment().Get(args.Input.LessonCommentId)
	if err != nil {
		return nil, err
	}

	lessonComment := lessonCommentPermit.Get()

	if args.Input.Body != nil {
		if err := lessonComment.Body.Set(args.Input.Body); err != nil {
			return nil, myerr.UnexpectedError{"failed to set lessonComment body"}
		}
	}

	lessonCommentPermit, err = r.Repos.LessonComment().Update(lessonComment)
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
	StudyId     string
}

func (r *RootResolver) UpdateStudy(
	ctx context.Context,
	args struct{ Input UpdateStudyInput },
) (*studyResolver, error) {
	studyPermit, err := r.Repos.Study().Get(args.Input.StudyId)
	if err != nil {
		return nil, err
	}

	study := studyPermit.Get()

	if args.Input.Description != nil {
		if err := study.Description.Set(args.Input.Description); err != nil {
			return nil, myerr.UnexpectedError{"failed to set study description"}
		}
	}
	if err := study.Id.Set(args.Input.StudyId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set study id"}
	}

	studyPermit, err = r.Repos.Study().Update(study)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: studyPermit, Repos: r.Repos}, nil
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
			topic, err := r.Repos.Topic().Create(t)
			if err != nil {
				return nil, err
			}
			topicId, err := topic.ID()
			if err != nil {
				return nil, err
			}
			topiced := &data.Topiced{}
			topiced.TopicId.Set(topicId)
			topiced.TopicableId.Set(args.Input.TopicableId)
			_, err = r.Repos.Topiced().Connect(topiced)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, t := range topics {
		name := t.Name.String
		if _, prs := newTopics[name]; !prs {
			topiced := &data.Topiced{}
			topiced.TopicId.Set(&t.Id)
			topiced.TopicableId.Set(topicableId)
			err := r.Repos.Topiced().Disconnect(topiced)
			if err != nil {
				return nil, err
			}
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

type UpdateUserInput struct {
	Bio         *string
	Login       *string
	Name        *string
	PublicEmail *string
	UserId      string
}

func (r *RootResolver) UpdateUser(
	ctx context.Context,
	args struct{ Input UpdateUserInput },
) (*userResolver, error) {
	userPermit, err := r.Repos.User().Get(args.Input.UserId)
	if err != nil {
		return nil, err
	}

	user := userPermit.Get()

	if args.Input.Bio != nil {
		if err := user.Bio.Set(args.Input.Bio); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user bio"}
		}
	}
	if args.Input.Login != nil {
		if err := user.Login.Set(args.Input.Login); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user login"}
		}
	}
	if args.Input.Name != nil {
		if err := user.Name.Set(args.Input.Name); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user name"}
		}
	}
	if args.Input.PublicEmail != nil {
		if err := user.PublicEmail.Set(args.Input.PublicEmail); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user public_email"}
		}
	}

	userPermit, err = r.Repos.User().Update(user)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: userPermit, Repos: r.Repos}, nil
}

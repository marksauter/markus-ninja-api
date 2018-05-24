package resolver

import (
	"context"
	"errors"
	"net"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
)

type AddUserEmailInput struct {
	Email string
}

func (r *RootResolver) AddUserEmail(
	ctx context.Context,
	args struct{ Input AddUserEmailInput },
) (*addUserEmailOutputResolver, error) {
	resolver := &addUserEmailOutputResolver{}

	viewer, ok := data.UserFromContext(ctx)
	if !ok {
		return resolver, errors.New("viewer not found")
	}

	userEmail := &data.UserEmail{}
	userEmail.EmailValue.Set(args.Input.Email)
	userEmail.UserId.Set(viewer.Id)

	userEmailPermit, err := r.Repos.UserEmail().Create(userEmail)
	if err != nil {
		return resolver, err
	}
	resolver.UserEmail = userEmailPermit

	evt := &data.EVT{}
	evt.EmailId.Set(userEmail.EmailId)
	evt.UserId.Set(viewer.Id)

	evtPermit, err := r.Repos.EVT().Create(evt)
	if err != nil {
		return resolver, err
	}
	resolver.EVT = evtPermit

	err = r.Svcs.Mail.SendEmailVerificationMail(
		args.Input.Email,
		viewer.Login.String,
		userEmail.EmailId.Short,
		evt.Token.String,
	)
	if err != nil {
		return resolver, err
	}

	return resolver, nil
}

type DeleteUserEmailInput struct {
	EmailId string
	UserId  string
}

type UpdateUserEmailInput struct {
	EmailId string
	Type    string
	UserId  string
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

	password, err := passwd.New(args.Input.Password)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create password")
		return nil, err
	}
	if err := password.CheckStrength(passwd.VeryWeak); err != nil {
		mylog.Log.WithError(err).Error("password failed strength check")
		return nil, err
	}

	if err := user.Login.Set(args.Input.Login); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user login"}
	}
	if err := user.Password.Set(password.Hash()); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user password"}
	}
	if err := user.PrimaryEmail.Value.Set(args.Input.Email); err != nil {
		return nil, myerr.UnexpectedError{"failed to set user primary_email"}
	}

	userPermit, err := r.Repos.User().Create(user)
	if err != nil {
		return nil, err
	}

	uResolver := &userResolver{User: userPermit, Repos: r.Repos}

	if user.Login.String != "guest" {
		evt := &data.EVT{}
		evt.EmailId.Set(user.PrimaryEmail.Id)
		evt.UserId.Set(user.Id)

		err = r.Svcs.EVT.Create(evt)
		if err != nil {
			return uResolver, err
		}

		err = r.Svcs.Mail.SendEmailVerificationMail(
			user.PrimaryEmail.Value.String,
			user.Login.String,
			user.PrimaryEmail.Id.String,
			evt.Token.String,
		)
		if err != nil {
			return uResolver, err
		}
	}

	return uResolver, nil
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

type UpdateUserInput struct {
	Bio    *string
	Email  *string
	Login  *string
	Name   *string
	UserId string
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
	if args.Input.Email != nil {
		if err := user.PublicEmail.Set(args.Input.Email); err != nil {
			return nil, myerr.UnexpectedError{"failed to set user public_email"}
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

	userPermit, err = r.Repos.User().Update(user)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: userPermit, Repos: r.Repos}, nil
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
	viewer, ok := data.UserFromContext(ctx)
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

type DeleteLessonInput struct {
	LessonId string
}

func (r *RootResolver) DeleteLesson(
	ctx context.Context,
	args struct{ Input DeleteLessonInput },
) (*graphql.ID, error) {
	lessonPermit, err := r.Repos.Lesson().Get(args.Input.LessonId)
	if err != nil {
		return nil, err
	}

	lesson := lessonPermit.Get()

	if err := lesson.Id.Set(args.Input.LessonId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set lesson id"}
	}

	if err := r.Repos.Lesson().Delete(lesson); err != nil {
		return nil, err
	}

	gqlID := graphql.ID(args.Input.LessonId)
	return &gqlID, nil
}

type UpdateLessonInput struct {
	Body     *string
	LessonId string
	Number   *int32
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
	if args.Input.Number != nil {
		if err := lesson.Number.Set(args.Input.Number); err != nil {
			return nil, myerr.UnexpectedError{"failed to set lesson number"}
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

type CreateStudyInput struct {
	Description *string
	Name        string
}

func (r *RootResolver) CreateStudy(
	ctx context.Context,
	args struct{ Input CreateStudyInput },
) (*studyResolver, error) {
	viewer, ok := data.UserFromContext(ctx)
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

type DeleteStudyInput struct {
	StudyId string
}

func (r *RootResolver) DeleteStudy(
	ctx context.Context,
	args struct{ Input DeleteStudyInput },
) (*graphql.ID, error) {
	studyPermit, err := r.Repos.Study().Get(args.Input.StudyId)
	if err != nil {
		return nil, err
	}

	study := studyPermit.Get()

	if err := study.Id.Set(args.Input.StudyId); err != nil {
		return nil, myerr.UnexpectedError{"failed to set study id"}
	}

	if err := r.Repos.Study().Delete(study); err != nil {
		return nil, err
	}

	gqlID := graphql.ID(args.Input.StudyId)
	return &gqlID, nil
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

type RequestEmailVerificationInput struct {
	Email string
}

func (r *RootResolver) RequestEmailVerification(
	ctx context.Context,
	args struct{ Input RequestEmailVerificationInput },
) (*evtResolver, error) {
	viewer, ok := data.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}

	userEmail, err := r.Repos.UserEmail().GetByEmail(
		args.Input.Email,
	)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, errors.New("`email` not found")
		}
		return nil, err
	}
	userId, err := userEmail.UserId()
	if err != nil {
		return nil, err
	}
	if viewer.Id.String != userId.String {
		return nil, errors.New("email already registered to another user")
	}

	isVerified, err := userEmail.IsVerified()
	if err != nil {
		return nil, err
	}
	if isVerified {
		return nil, errors.New("email already verified")
	}

	emailId, err := userEmail.EmailId()
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

	resolver := &evtResolver{EVT: evtPermit, Repos: r.Repos}

	emailValue, err := userEmail.EmailValue()
	if err != nil {
		return nil, err
	}
	userLogin, err := userEmail.UserLogin()
	if err != nil {
		return nil, err
	}

	err = r.Svcs.Mail.SendEmailVerificationMail(
		emailValue,
		userLogin,
		emailId.Short,
		evt.Token.String,
	)
	if err != nil {
		return resolver, err
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
	userEmail, err := r.Svcs.UserEmail.GetByEmail(args.Input.Email)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, errors.New("`email` not found")
		}
		return nil, err
	}

	requestIp, ok := ctx.Value("requester_ip").(*net.IPNet)
	if !ok {
		return nil, errors.New("requester ip not found")
	}

	prt := &data.PRT{}
	if err := prt.EmailId.Set(userEmail.EmailId); err != nil {
		mylog.Log.Error(err)
		return nil, myerr.UnexpectedError{"failed to set prt email_id"}
	}
	if err := prt.UserId.Set(userEmail.UserId); err != nil {
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

	err = r.Svcs.Mail.SendPasswordResetMail(
		args.Input.Email,
		userEmail.UserLogin.String,
		prt.Token.String,
	)
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

	prt, err := r.Svcs.PRT.GetByPK(user.Id.String, args.Input.Token)
	if err != nil {
		return false, err
	}

	if prt.ExpiresAt.Time.Before(time.Now()) {
		return false, errors.New("token has expired")
	}

	if prt.EndedAt.Status == pgtype.Present {
		return false, errors.New("token has already ended")
	}

	password, err := passwd.New(args.Input.Password)
	if err != nil {
		return false, err
	}
	if err := password.CheckStrength(passwd.VeryWeak); err != nil {
		return false, err
	}

	err = user.Password.Set(password.Hash())
	if err != nil {
		return false, myerr.UnexpectedError{"failed to set user password"}
	}

	if err := r.Svcs.User.Update(user); err != nil {
		return false, myerr.UnexpectedError{"failed to update user"}
	}

	endIp, ok := ctx.Value("requester_ip").(*net.IPNet)
	if !ok {
		return false, errors.New("requester ip not found")
	}

	if err := prt.EndIP.Set(endIp); err != nil {
		return false, myerr.UnexpectedError{"failed to set prt end_ip"}
	}
	if err := prt.EndedAt.Set(time.Now()); err != nil {
		return false, myerr.UnexpectedError{"failed to set prt ended_at"}
	}

	if err := r.Svcs.PRT.Update(prt); err != nil {
		return false, myerr.UnexpectedError{"failed to update prt"}
	}

	return true, nil
}

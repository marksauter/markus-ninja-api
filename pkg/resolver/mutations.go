package resolver

import (
	"context"
	"errors"
	"net"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
)

type AddUserEmailInput struct {
	Email  string
	UserId string
}

// func (r *RootResolver) AddUserEmail(
//   ctx context.Context,
//   args struct{ Input AddUserEmailInput },
// ) (*userEmailResolver, error) {
//   email := data.Email{}
//   email.Value.Set(args.Input.Email)
//   err = r.Repos.Email().Create(email)
//   if err != nil {
//     return nil, err
//   }
//
//   user := data.User{}
//   err = r.Repos.User().Get(args.Input.UserId)
//   if err != nil {
//     return nil, err
//   }
//
//   userEmail := data.UserEmail{}
//   userEmail.EmailId.Set(email.Id)
//   userEmail.UserId.Set(user.Id)
//
//   err = r.Repos.UserEmail().Create(userEmail)
//   if err != nil {
//     return nil, err
//   }
//
// }

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
	var user data.User

	password, err := passwd.New(args.Input.Password)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create password")
		return nil, err
	}
	if err := password.CheckStrength(passwd.VeryWeak); err != nil {
		mylog.Log.WithError(err).Error("password failed strength check")
		return nil, err
	}

	user.Login.Set(args.Input.Login)
	user.Password.Set(password.Hash())
	user.PrimaryEmail.Value.Set(args.Input.Email)

	userPermit, err := r.Repos.User().Create(&user)
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
	id, err := oid.Parse(args.Input.UserId)
	if err != nil {
		return nil, err
	}

	user := &data.User{Id: *id}

	err = r.Repos.User().Delete(user)
	if err != nil {
		return nil, err
	}
	gqlID := graphql.ID(args.Input.UserId)

	return &gqlID, nil
}

type UpdateUserInput struct {
	Bio    *string
	UserId string
	Login  *string
	Name   *string
}

func (r *RootResolver) UpdateUser(
	ctx context.Context,
	args struct{ Input UpdateUserInput },
) (*userResolver, error) {
	var user data.User

	id, err := oid.Parse(args.Input.UserId)
	if err != nil {
		return nil, err
	}
	user.Id.Set(id.String)

	if args.Input.Bio != nil {
		user.Bio.Set(args.Input.Bio)
	}
	if args.Input.Login != nil {
		user.Login.Set(args.Input.Login)
	}
	if args.Input.Name != nil {
		user.Name.Set(args.Input.Name)
	}

	userPermit, err := r.Repos.User().Update(&user)
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

	var lesson data.Lesson

	lesson.Body.Set(args.Input.Body)
	lesson.StudyId.Set(args.Input.StudyId)
	lesson.Title.Set(args.Input.Title)
	lesson.UserId.Set(viewer.Id.String)

	lessonPermit, err := r.Repos.Lesson().Create(&lesson)
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

	var study data.Study

	if args.Input.Description != nil {
		study.Description.Set(args.Input.Description)
	}
	study.Name.Set(args.Input.Name)
	study.UserId.Set(viewer.Id.String)

	studyPermit, err := r.Repos.Study().Create(&study)
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
	if viewer.Id.String != userId {
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
	evt.EmailId.Set(emailId)
	evt.UserId.Set(userId)

	evtPermit, err := r.Repos.EVT().Create(evt)
	if err != nil {
		return nil, err
	}

	resolver := &evtResolver{EVT: evtPermit, Repos: r.Repos}

	err = r.Svcs.Mail.SendEmailVerificationMail(
		userEmail.EmailValue(),
		userEmail.UserLogin(),
		emailId,
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
	user, err := r.Svcs.User.GetCredentialsByEmail(args.Input.Email)
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
	prt.Email.Set(args.Input.Email)
	prt.UserId.Set(user.Id)
	err = prt.RequestIP.Set(requestIp)
	if err != nil {
		return nil, err
	}

	prtPermit, err := r.Repos.PRT().Create(prt)
	if err != nil {
		return nil, err
	}

	resolver := &prtResolver{PRT: prtPermit, Repos: r.Repos}

	err = r.Svcs.Mail.SendPasswordResetMail(
		args.Input.Email,
		user.Login.String,
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
	prt, err := r.Svcs.PRT.GetByPK(args.Input.Token)
	if err != nil {
		return false, err
	}
	return false, nil
}

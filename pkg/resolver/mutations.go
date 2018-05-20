package resolver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type AddUserEmailInput struct {
	Email  string
	UserId string
}

func (r *RootResolver) AddUserEmail(
	ctx context.Context,
	args struct{ Input AddUserEmailInput },
) (*userEmailResolver, error) {
	fields, err := r.Repos.UserEmail().AddPermission(perm.CreateUserEmail)
	if err != nil {
		return nil, err
	}

	email := data.Email{}
	email.Value.Set(args.Input.Email)
	err = r.Repos.Email().Create(email)
	if err != nil {
		return nil, err
	}

	user := data.User{}
	err = r.Repos.User().Get(args.Input.UserId)
	if err != nil {
		return nil, err
	}

	userEmail := data.UserEmail{}
	userEmail.EmailId.Set(email.Id)
	userEmail.UserId.Set(user.Id)

	err = r.Repos.UserEmail().Create(userEmail)
	if err != nil {
		return nil, err
	}

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
	fields, err := r.Repos.User().AddPermission(perm.CreateUser)
	if err != nil {
		return nil, err
	}

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
	user.PrimaryEmail.Set(args.Input.Email)

	_, err = r.Repos.User().Create(&user)
	if err == repo.ErrAccessDenied {
		return nil, fmt.Errorf(
			`access_denied: you may only include %s`,
			strings.Join(fields, ", "),
		)
	}
	if err != nil {
		return nil, err
	}

	node, err := r.Node(ctx, struct{ Id string }{Id: user.Id.String})
	if err != nil {
		return nil, err
	}
	uResolver, ok := node.ToUser()
	if !ok {
		return nil, myerr.UnexpectedError{"failed to convert nodeResolver to userResolver"}
	}

	if user.Login.String != "guest" {
		avt := &data.EmailVerificationTokenModel{}
		avt.UserId.Set(user.Id.String)

		err = r.Svcs.AVT.Create(avt)
		if err != nil {
			return uResolver, err
		}

		err = r.Svcs.Mail.SendEmailVerificationMail(
			user.PrimaryEmail.String,
			user.Login.String,
			avt.Token.String,
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
) (*string, error) {
	_, err := r.Repos.User().AddPermission(perm.DeleteUser)
	if err != nil {
		return nil, err
	}

	id, err := oid.Parse(args.Input.UserId)
	if err != nil {
		return nil, err
	}

	err = r.Repos.User().Delete(id.String)
	if err != nil {
		return nil, err
	}

	return &args.Input.UserId, nil
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
	_, err := r.Repos.User().AddPermission(perm.UpdateUser)
	if err != nil {
		return nil, err
	}

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
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()

	fields, err := r.Repos.Lesson().AddPermission(perm.CreateLesson)
	if err != nil {
		return nil, err
	}

	var lesson data.Lesson

	lesson.Body.Set(args.Input.Body)
	lesson.StudyId.Set(args.Input.StudyId)
	lesson.Title.Set(args.Input.Title)
	lesson.UserId.Set(viewerId)

	_, err = r.Repos.Lesson().Create(&lesson)
	if err == repo.ErrAccessDenied {
		return nil, fmt.Errorf(
			`access_denied: you may only include %s`,
			strings.Join(fields, ", "),
		)
	}
	if err != nil {
		return nil, err
	}

	node, err := r.Node(ctx, struct{ Id string }{Id: lesson.Id.String})
	if err != nil {
		return nil, err
	}
	resolver, ok := node.ToLesson()
	if !ok {
		return nil, myerr.UnexpectedError{"failed to convert nodeResolver to lessonResolver"}
	}

	return resolver, nil
}

type CreateStudyInput struct {
	Description *string
	Name        string
}

func (r *RootResolver) CreateStudy(
	ctx context.Context,
	args struct{ Input CreateStudyInput },
) (*studyResolver, error) {
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()

	fields, err := r.Repos.Study().AddPermission(perm.CreateStudy)
	if err != nil {
		return nil, err
	}

	var study data.Study

	if args.Input.Description != nil {
		study.Description.Set(args.Input.Description)
	}
	study.Name.Set(args.Input.Name)
	study.UserId.Set(viewerId)

	_, err = r.Repos.Study().Create(&study)
	if err == repo.ErrAccessDenied {
		return nil, fmt.Errorf(
			`access_denied: you may only include %s`,
			strings.Join(fields, ", "),
		)
	}
	if err != nil {
		return nil, err
	}

	node, err := r.Node(ctx, struct{ Id string }{Id: study.Id.String})
	if err != nil {
		return nil, err
	}
	resolver, ok := node.ToStudy()
	if !ok {
		return nil, myerr.UnexpectedError{"failed to convert nodeResolver to studyResolver"}
	}

	return resolver, nil
}

package data_test

import (
	"testing"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
)

func newUser() *data.UserModel {
	user := data.UserModel{}
	user.Email.Set("test@example.com")
	user.Login.Set("test")
	user.Password.Set([]byte("password"))
	return &user
}

func TestDataUsersLifeCycle(t *testing.T) {
	db := mydb.NewTestDB(t)
	defer db.Close()
	userSvc := data.NewUserService(db.DB)

	input := newUser()
	err := userSvc.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	userId := input.Id.String

	user, err := userSvc.GetByLogin(input.Login.String)
	if err != nil {
		t.Fatal(err)
	}
	if user.Id.String != userId {
		t.Errorf("Expected %v, got %v", userId, user.Id)
	}
}

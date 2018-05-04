package data_test

import (
	"bytes"
	"testing"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
)

func newUser() *data.UserModel {
	user := data.UserModel{}
	user.Login.Set("test")
	user.Password.Set([]byte("password"))
	user.PrimaryEmail.Set("test@example.com")
	return &user
}

func TestDataUsersLifeCycle(t *testing.T) {
	db := mydb.NewTestDB(t)
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
	if bytes.Compare(user.Password.Bytes, input.Password.Bytes) != 0 {
		t.Errorf("Expected %v, got %v", input.Password, user.Password)
	}
	if user.PrimaryEmail != input.PrimaryEmail {
		t.Errorf("Expected %v, got %v", input.PrimaryEmail, user.PrimaryEmail)
	}

	// user, err = userSvc.GetByPrimaryEmail(input.PrimaryEmail.String)
	// if err != nil {
	//   t.Fatal(err)
	// }
	// if user.Id.String != userId {
	//   t.Errorf("Expected %v, got %v", userId, user.Id)
	// }
	// if bytes.Compare(user.Password.Bytes, input.Password.Bytes) != 0 {
	//   t.Errorf("Expected %v, got %v", input.Password, user.Password)
	// }
	// if user.PrimaryEmail != input.PrimaryEmail {
	//   t.Errorf("Expected %v, got %v", input.PrimaryEmail, user.PrimaryEmail)
	// }

	user, err = userSvc.GetById(input.Id.String)
	if err != nil {
		t.Fatal(err)
	}
	if user.Id.String != userId {
		t.Errorf("Expected %v, got %v", userId, user.Id)
	}
	if bytes.Compare(user.Password.Bytes, input.Password.Bytes) != 0 {
		t.Errorf("Expected %v, got %v", input.Password, user.Password)
	}
	if user.PrimaryEmail != input.PrimaryEmail {
		t.Errorf("Expected %v, got %v", input.PrimaryEmail, user.PrimaryEmail)
	}
}

func TestDataCreateUserHandlesLoginUniqueness(t *testing.T) {
	db := mydb.NewTestDB(t)
	userSvc := data.NewUserService(db.DB)

	user := newUser()
	err := userSvc.Create(user)
	if err != nil {
		t.Fatal(err)
	}

	user = newUser()
	actual := userSvc.Create(user)
	expected := data.DuplicateFieldError("login")
	if actual != expected {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// func TestDataCreateUserHandlesPrimaryEmailUniqueness(t *testing.T) {
//   db := mydb.NewTestDB(t)
//   userSvc := data.NewUserService(db.DB)
//
//   user := newUser()
//   user.Email.Set("test@example.com")
//   err := userSvc.Create(user)
//   if err != nil {
//     t.Fatal(err)
//   }
//
//   user.Login.Set("otherlogin")
//   actual := userSvc.Create(user)
//   expected := data.DuplicateFieldError("primary_email")
//   if actual != expected {
//     t.Fatalf("Expected %v, got %v", expected, actual)
//   }
// }

func BenchmarkDataGetUser(b *testing.B) {
	db := mydb.NewTestDB(b)
	userSvc := data.NewUserService(db.DB)

	user := newUser()
	err := userSvc.Create(user)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := userSvc.GetById(user.Id.String)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataGetUserByLogin(b *testing.B) {
	db := mydb.NewTestDB(b)
	userSvc := data.NewUserService(db.DB)

	user := newUser()
	err := userSvc.Create(user)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := userSvc.GetByLogin(user.Login.String)
		if err != nil {
			b.Fatal(err)
		}
	}
}

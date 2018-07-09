package data_test

import (
	"bytes"
	"testing"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
)

func newUser() *data.User {
	user := data.User{}
	user.Login.Set("test")
	user.Password.Set([]byte("password"))
	user.PrimaryEmail.Set("test@example.com")
	return &user
}

func TestDataUsersLifeCycle(t *testing.T) {
	db := mydb.NewTestDB(t)

	input := newUser()
	output, err := data.CreateUser(db, input)
	if err != nil {
		t.Fatal(err)
	}
	userId := output.Id.String

	user, err := data.GetUserCredentialsByLogin(db, input.Login.String)
	if err != nil {
		t.Fatal(err)
	}
	if user.Id.String != userId {
		t.Errorf("Expected %v, got %v", userId, user.Id)
	}
	if bytes.Compare(user.Password.Bytes, input.Password.Bytes) != 0 {
		t.Errorf("Expected %v, got %v", input.Password.Bytes, user.Password.Bytes)
	}
	if user.PrimaryEmail.String != input.PrimaryEmail.String {
		t.Errorf("Expected %v, got %v", input.PrimaryEmail.String, user.PrimaryEmail.String)
	}

	user, err = data.GetUser(db, output.Id.String)
	if err != nil {
		t.Fatal(err)
	}
	if user.Id.String != userId {
		t.Errorf("Expected %v, got %v", userId, user.Id)
	}
	if user.Login.String != input.Login.String {
		t.Errorf("Expected %v, got %v", input.Login.String, user.Login.String)
	}
}

func TestDataCreateUserHandlesLoginUniqueness(t *testing.T) {
	db := mydb.NewTestDB(t)

	input := newUser()
	_, err := data.CreateUser(db, input)
	if err != nil {
		t.Fatal(err)
	}

	input = newUser()
	_, actual := data.CreateUser(db, input)
	expected := data.DuplicateFieldError("login")
	if actual != expected {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func TestDataCreateUserHandlesPrimaryEmailUniqueness(t *testing.T) {
	db := mydb.NewTestDB(t)

	user := newUser()
	user.PrimaryEmail.Set("test@example.com")
	_, err := data.CreateUser(db, user)
	if err != nil {
		t.Fatal(err)
	}

	user.Login.Set("otherlogin")
	_, actual := data.CreateUser(db, user)
	expected := data.DuplicateFieldError("primary_email")
	if actual != expected {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func BenchmarkDataGetUser(b *testing.B) {
	db := mydb.NewTestDB(b)

	user := newUser()
	output, err := data.CreateUser(db, user)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := data.GetUser(db, output.Id.String)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataGetUserByLogin(b *testing.B) {
	db := mydb.NewTestDB(b)

	input := newUser()
	_, err := data.CreateUser(db, input)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := data.GetUserByLogin(db, input.Login.String)
		if err != nil {
			b.Fatal(err)
		}
	}
}

package model

type User struct {
	id string
}

func (u *User) Id() string {
	return u.id
}

type NewUserInput struct {
	Id string
}

func NewUser(input *NewUserInput) *User {
	return &User{id: input.Id}
}

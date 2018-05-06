package model

import "time"

// The following types are for internal use only.
// These types represent the permissable fields for each corresponding data
// type.

type User struct {
	Bio       string    `perm:"public"`
	CreatedAt time.Time `perm:"private"`
	Email     string    `perm:"public"`
	Id        string    `perm:"public"`
	Login     string    `perm:"public"`
	Name      string    `perm:"public"`
	UpdatedAt time.Time `perm:"private"`
}

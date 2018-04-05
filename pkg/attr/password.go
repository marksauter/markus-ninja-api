package attr

import (
	"log"

	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	value []byte
}

func NewPassword(password string) *Password {
	return &Password{value: []byte(password)}
}

func (p *Password) Hash() ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(p.value, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (p *Password) CompareToHash(hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, p.value)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (p *Password) CheckStrength() int {
	entropy := zxcvbn.PasswordStrength(string(p.value), nil)
	return entropy.Score
}

package passwd

import (
	"log"

	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

type PasswordStrength int

const (
	VeryWeak PasswordStrength = iota
	Weak
	Moderate
	Strong
	VeryStrong
)

type Password struct {
	hash     []byte
	strength PasswordStrength
	value    []byte
}

func New(password string) *Password {
	return &Password{value: []byte(password)}
}

func (p *Password) Hash() ([]byte, error) {
	if p.hash == nil {
		hash, err := bcrypt.GenerateFromPassword(p.value, bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		p.hash = hash
	}
	return p.hash, nil
}

func (p *Password) CompareToHash(hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, p.value)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (p *Password) CheckStrength(s PasswordStrength) bool {
	entropy := zxcvbn.PasswordStrength(string(p.value), nil)
	p.strength = PasswordStrength(entropy.Score)
	return p.strength >= s
}

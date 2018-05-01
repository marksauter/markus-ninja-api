package passwd

import (
	"errors"

	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

var ErrTooWeak = errors.New("password too weak")

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

func (p *Password) String() string {
	return "*****"
}

func (p *Password) Hash() []byte {
	if p.hash == nil {
		hash, err := bcrypt.GenerateFromPassword(p.value, bcrypt.DefaultCost)
		if err != nil {
			// This should never happen, so panic if it does
			panic(err)
		}
		p.hash = hash
	}
	return p.hash
}

func (p *Password) CompareToHash(hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, p.value)
}

func (p *Password) CheckStrength(s PasswordStrength) error {
	entropy := zxcvbn.PasswordStrength(string(p.value), nil)
	p.strength = PasswordStrength(entropy.Score)
	if p.strength < s {
		return ErrTooWeak
	}
	return nil
}

// This doesn't work for some reason...
// func (p *Password) Value() (driver.Value, error) {
//   return p.Hash()
// }

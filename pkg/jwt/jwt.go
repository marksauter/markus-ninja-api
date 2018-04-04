package jwt

type Jwt interface {
	Encrypt(*Payload) *Token
	Decrypt(*Token) (*Payload, error)
}

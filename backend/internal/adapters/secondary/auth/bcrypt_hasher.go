// internal/adapters/secondary/auth/bcrypt_hasher.go
package auth

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
    return &BcryptHasher{}
}

func (b *BcryptHasher) Compare(hash string, password string) error {
    return bcrypt.CompareHashAndPassword(
        []byte(hash),
        []byte(password),
    )
}

package password_hash

import "golang.org/x/crypto/bcrypt"

// PasswordHasher Интерфейс стратегии хеширования паролей
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// BcryptHasher Конкретная стратегия - алгоритм bcrypt
type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

func (b *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	return string(bytes), err
}

func (b *BcryptHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

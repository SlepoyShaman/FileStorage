package user_service

import (
	"os"

	"github.com/SlepoyShaman/FileStorage/services/password_hash"
)

// UserService Сервис, который использует стратегию хеширования
type UserService struct {
	hasher password_hash.PasswordHasher
}

func NewUserService() *UserService {
	return &UserService{
		// Конкретная реализация выбрана здесь один раз и навсегда
		hasher: password_hash.NewBcryptHasher(os.Getenv("default_hash_cost")),
	}
}

func (s *UserService) SetPassword(user *User, plainPassword string) error {
	hashed, err := s.hasher.Hash(plainPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hashed
	return nil
}

func (s *UserService) ValidatePassword(user *User, plainPassword string) bool {
	return s.hasher.Verify(plainPassword, user.PasswordHash)
}

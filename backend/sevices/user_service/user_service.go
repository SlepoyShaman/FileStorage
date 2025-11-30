package user_service

// UserService Сервис, который использует стратегию хеширования
type UserService struct {
	hasher PasswordHasher
}

func NewUserService(hasher PasswordHasher) *UserService {
	return &UserService{hasher: hasher}
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

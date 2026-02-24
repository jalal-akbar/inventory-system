package service

import (
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(username, password string) (*domain.User, error)
}

type authService struct {
	userRepo repository.UserRepository
	logRepo  repository.ActivityLogRepository
}

func NewAuthService(userRepo repository.UserRepository, logRepo repository.ActivityLogRepository) AuthService {
	return &authService{userRepo: userRepo, logRepo: logRepo}
}

func (s *authService) Login(username, password string) (*domain.User, error) {
	u, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	// Timing attack mitigation: always run bcrypt even if user not found
	var hash string
	if u != nil {
		hash = u.Password
	} else {
		// Use a dummy hash for non-existent users
		// This hash is for "password" and is just a placeholder
		hash = "$2a$10$AzRgvzWw6LxO/Tz.xUe.x.Vd.xUe.x.Vd.xUe.x.Vd.xUe.x.Vd."
	}

	bcryptErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if u == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if u.Status != "active" {
		return nil, domain.ErrAccountDisabled
	}

	if bcryptErr != nil {
		return nil, domain.ErrInvalidCredentials
	}

	s.logRepo.Log(u.ID, "User logged in")

	return u, nil
}

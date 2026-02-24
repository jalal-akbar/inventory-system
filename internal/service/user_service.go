package service

import (
	"errors"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UpdateUserInput struct {
	Username string
	Role     string
	Status   string
	Password string
	Language string
}

type UserService interface {
	GetAll() ([]domain.User, error)
	Create(username, password, role string) error
	Update(id int, input UpdateUserInput) error
	SetStatus(id int, status string) error
	UpdateLanguage(id int, lang string) error
	GetUserCount() (int, error)
	ChangePassword(id int, oldPassword, newPassword string) error
}

type userService struct {
	userRepo repository.UserRepository
	logRepo  repository.ActivityLogRepository
}

func NewUserService(userRepo repository.UserRepository, logRepo repository.ActivityLogRepository) UserService {
	return &userService{userRepo: userRepo, logRepo: logRepo}
}

func (s *userService) GetAll() ([]domain.User, error) {
	return s.userRepo.GetAll()
}

func (s *userService) Create(username, password, role string) error {
	if username == "" || password == "" || role == "" {
		return errors.New("all fields are required")
	}

	existing, _ := s.userRepo.FindByUsername(username)
	if existing != nil {
		return errors.New("username already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u := &domain.User{
		Username: username,
		Password: string(hashed),
		Role:     role,
		Status:   "active",
	}

	return s.userRepo.Create(u)
}

func (s *userService) Update(id int, input UpdateUserInput) error {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user not found")
	}

	u.Username = input.Username
	u.Role = input.Role
	u.Status = input.Status
	if input.Language != "" {
		u.Language = input.Language
	}

	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashed)
	}

	return s.userRepo.Update(u)
}

func (s *userService) SetStatus(id int, status string) error {
	return s.userRepo.SetStatus(id, status)
}

func (s *userService) UpdateLanguage(id int, lang string) error {
	return s.userRepo.UpdateLanguage(id, lang)
}

func (s *userService) GetUserCount() (int, error) {
	return s.userRepo.Count()
}

func (s *userService) ChangePassword(id int, oldPassword, newPassword string) error {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("invalid current password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashed)
	return s.userRepo.Update(u)
}

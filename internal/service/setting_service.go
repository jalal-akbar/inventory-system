package service

import (
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

type SettingService interface {
	GetSettings() (*domain.Setting, error)
	UpdateGlobalSettings(userID int, businessName, address, phone, currencySymbol, timezone string) error
}

type settingService struct {
	settingRepo repository.SettingRepository
	logRepo     repository.ActivityLogRepository
}

func NewSettingService(settingRepo repository.SettingRepository, logRepo repository.ActivityLogRepository) SettingService {
	return &settingService{settingRepo: settingRepo, logRepo: logRepo}
}

func (s *settingService) GetSettings() (*domain.Setting, error) {
	return s.settingRepo.GetSettings()
}

func (s *settingService) UpdateGlobalSettings(userID int, businessName, address, phone, currencySymbol, timezone string) error {
	err := s.settingRepo.UpdateSettings(businessName, address, phone, currencySymbol, timezone)
	if err == nil {
		s.logRepo.Log(userID, "Updated global settings (Name: "+businessName+")")
	}
	return err
}

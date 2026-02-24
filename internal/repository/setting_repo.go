package repository

import (
	"database/sql"
	"inventory-system/internal/domain"
)

type SettingRepository interface {
	GetSettings() (*domain.Setting, error)
	UpdateSettings(businessName, address, phone, currencySymbol, timezone string) error
}

type mysqlSettingRepository struct {
	db *sql.DB
}

func NewSettingRepository(db *sql.DB) SettingRepository {
	return &mysqlSettingRepository{db: db}
}

func (r *mysqlSettingRepository) GetSettings() (*domain.Setting, error) {
	s := &domain.Setting{}
	err := r.db.QueryRow("SELECT id, business_name, COALESCE(address, ''), COALESCE(phone, ''), COALESCE(currency_symbol, 'Rp'), timezone, updated_at FROM settings LIMIT 1").
		Scan(&s.ID, &s.BusinessName, &s.Address, &s.Phone, &s.CurrencySymbol, &s.Timezone, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *mysqlSettingRepository) UpdateSettings(businessName, address, phone, currencySymbol, timezone string) error {
	_, err := r.db.Exec("UPDATE settings SET business_name = ?, address = ?, phone = ?, currency_symbol = ?, timezone = ?, updated_at = CURRENT_TIMESTAMP WHERE id = 1",
		businessName, address, phone, currencySymbol, timezone)
	return err
}

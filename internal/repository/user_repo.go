package repository

import (
	"database/sql"
	"inventory-system/internal/domain"
)

type UserRepository interface {
	FindByID(id int) (*domain.User, error)
	FindByUsername(username string) (*domain.User, error)
	GetAll() ([]domain.User, error)
	Create(u *domain.User) error
	Update(u *domain.User) error
	UpdateLanguage(id int, lang string) error
	SetStatus(id int, status string) error
	Count() (int, error)
}

type mysqlUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepository{db: db}
}

func (r *mysqlUserRepository) FindByID(id int) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow("SELECT id, username, password, role, status, language, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.Language, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *mysqlUserRepository) FindByUsername(username string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow("SELECT id, username, password, role, status, language, created_at FROM users WHERE username = ?", username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.Language, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *mysqlUserRepository) GetAll() ([]domain.User, error) {
	rows, err := r.db.Query("SELECT id, username, password, role, status, language, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.Language, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *mysqlUserRepository) Create(u *domain.User) error {
	_, err := r.db.Exec("INSERT INTO users (username, password, role, status, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)",
		u.Username, u.Password, u.Role, u.Status)
	return err
}

func (r *mysqlUserRepository) Update(u *domain.User) error {
	_, err := r.db.Exec("UPDATE users SET username = ?, role = ?, status = ?, language = ?, password = ? WHERE id = ?",
		u.Username, u.Role, u.Status, u.Language, u.Password, u.ID)
	return err
}

func (r *mysqlUserRepository) UpdateLanguage(id int, lang string) error {
	_, err := r.db.Exec("UPDATE users SET language = ? WHERE id = ?", lang, id)
	return err
}

func (r *mysqlUserRepository) SetStatus(id int, status string) error {
	_, err := r.db.Exec("UPDATE users SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *mysqlUserRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&count)
	return count, err
}

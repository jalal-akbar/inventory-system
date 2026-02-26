package repository

import (
	"database/sql"
	"inventory-system/internal/domain"
)

type ActivityLogRepository interface {
	Log(userID int, action string) error
	GetLatest(limit int) ([]domain.ActivityLog, error)
	Search(startDate, endDate, sort string) ([]domain.ActivityLog, error)
	WithTx(tx *sql.Tx) ActivityLogRepository
}

type mysqlActivityLogRepository struct {
	db DBExecutor
}

func NewActivityLogRepository(db *sql.DB) ActivityLogRepository {
	return &mysqlActivityLogRepository{db: db}
}

func (r *mysqlActivityLogRepository) getDB() DBExecutor {
	return r.db
}

func (r *mysqlActivityLogRepository) WithTx(tx *sql.Tx) ActivityLogRepository {
	return &mysqlActivityLogRepository{db: tx}
}

func (r *mysqlActivityLogRepository) Log(userID int, action string) error {
	_, err := r.getDB().Exec("INSERT INTO activity_logs (user_id, action, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)", userID, action)
	return err
}

func (r *mysqlActivityLogRepository) GetLatest(limit int) ([]domain.ActivityLog, error) {
	rows, err := r.getDB().Query(`
		SELECT al.id, al.user_id, u.username, al.action, al.created_at 
		FROM activity_logs al 
		LEFT JOIN users u ON al.user_id = u.id 
		ORDER BY al.created_at DESC 
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.ActivityLog
	for rows.Next() {
		var l domain.ActivityLog
		var username sql.NullString
		if err := rows.Scan(&l.ID, &l.UserID, &username, &l.Action, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.Username = "Unknown"
		if username.Valid {
			l.Username = username.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *mysqlActivityLogRepository) Search(startDate, endDate, sort string) ([]domain.ActivityLog, error) {
	query := "SELECT al.id, al.user_id, u.username, al.action, al.created_at FROM activity_logs al LEFT JOIN users u ON al.user_id = u.id"
	var params []interface{}

	if startDate != "" && endDate != "" {
		query += " WHERE DATE(al.created_at) BETWEEN ? AND ?"
		params = append(params, startDate, endDate)
	} else if startDate != "" {
		query += " WHERE DATE(al.created_at) = ?"
		params = append(params, startDate)
	}

	if sort != "ASC" {
		sort = "DESC"
	}
	query += " ORDER BY al.created_at " + sort

	rows, err := r.getDB().Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.ActivityLog
	for rows.Next() {
		var l domain.ActivityLog
		var username sql.NullString
		if err := rows.Scan(&l.ID, &l.UserID, &username, &l.Action, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.Username = "Unknown"
		if username.Valid {
			l.Username = username.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

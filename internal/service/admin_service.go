package service

import (
	"database/sql"
	"fmt"
	"inventory-system/internal/repository"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SystemHealth struct {
	Status         string                 `json:"status"`
	Message        string                 `json:"message"`
	DiskFree       string                 `json:"disk_free"`
	DiskTotal      string                 `json:"disk_total"`
	DiskUsagePct   float64                `json:"disk_usage_pct"`
	MemoryUsed     string                 `json:"memory_used"`
	MemoryTotal    string                 `json:"memory_total"`
	MemoryUsagePct float64                `json:"memory_usage_pct"`
	DBStatus       string                 `json:"db_status"`
	Uptime         string                 `json:"uptime"`
	Checks         map[string]HealthCheck `json:"checks"`
}

type AdminService interface {
	ApproveItem(id int, itemType string, adminID int) error
	RejectItem(id int, itemType string, adminID int) error
	ApproveGroup(productID int, adminID int) error
	ApproveAll(adminID int) error
	GetSystemHealth() (*SystemHealth, error)
	CreateBackup() (string, error)
	GetPendingItems() ([]map[string]interface{}, error)
}

type adminService struct {
	db          *sql.DB
	productRepo repository.ProductRepository
	batchRepo   repository.ProductBatchRepository
	entryRepo   repository.StockEntryRepository
	saleRepo    repository.SaleRepository
	logRepo     repository.ActivityLogRepository
	startTime   time.Time
}

func NewAdminService(db *sql.DB, pRepo repository.ProductRepository, bRepo repository.ProductBatchRepository, eRepo repository.StockEntryRepository, sRepo repository.SaleRepository, lRepo repository.ActivityLogRepository) AdminService {
	return &adminService{
		db:          db,
		productRepo: pRepo,
		batchRepo:   bRepo,
		entryRepo:   eRepo,
		saleRepo:    sRepo,
		logRepo:     lRepo,
		startTime:   time.Now(),
	}
}

func (s *adminService) ApproveItem(id int, itemType string, adminID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	switch itemType {
	case "product", "drug":
		if err := s.productRepo.Verify(id); err != nil {
			return err
		}
		if err := s.batchRepo.VerifyByProduct(id); err != nil {
			return err
		}
		if err := s.entryRepo.VerifyByProduct(id); err != nil {
			return err
		}
		s.logRepo.Log(adminID, fmt.Sprintf("Verified product group for ID: %d", id))

	case "batch":
		if err := s.batchRepo.Verify(id); err != nil {
			return err
		}
		if err := s.entryRepo.VerifyByBatch(id); err != nil {
			return err
		}
		s.logRepo.Log(adminID, fmt.Sprintf("Verified batch ID: %d", id))

	case "stock_entry":
		if err := s.entryRepo.Verify(id); err != nil {
			return err
		}
		s.logRepo.Log(adminID, fmt.Sprintf("Verified stock entry ID: %d", id))

	case "void_request":
		if err := s.saleRepo.SetStatus(id, "void"); err != nil {
			return err
		}
		items, err := s.saleRepo.GetSaleItems(id)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := s.batchRepo.UpdateStock(item.BatchID, item.Quantity); err != nil {
				return err
			}
		}
		s.logRepo.Log(adminID, fmt.Sprintf("Approved void request for sale ID: %d", id))

	default:
		return fmt.Errorf("unknown item type: %s", itemType)
	}

	return tx.Commit()
}

func (s *adminService) RejectItem(id int, itemType string, adminID int) error {
	switch itemType {
	case "stock_entry":
		if err := s.entryRepo.Reject(id); err != nil {
			return err
		}
		s.logRepo.Log(adminID, fmt.Sprintf("Rejected stock entry ID: %d", id))

	case "void_request":
		if err := s.saleRepo.SetStatus(id, "active"); err != nil {
			return err
		}
		s.logRepo.Log(adminID, fmt.Sprintf("Rejected void request for sale ID: %d", id))

	default:
		// For others, we might just keep them unverified or delete
		return fmt.Errorf("rejection for type %s not fully implemented", itemType)
	}
	return nil
}

func (s *adminService) ApproveGroup(productID int, adminID int) error {
	return s.ApproveItem(productID, "product", adminID)
}

func (s *adminService) ApproveAll(adminID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE products SET is_verified = 1 WHERE is_verified = 0"); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE product_batches SET is_verified = 1 WHERE is_verified = 0"); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE stock_entries SET is_verified = 1 WHERE is_verified = 0 AND status = 'approved'"); err != nil {
		return err
	}

	s.logRepo.Log(adminID, "Bulk approved all pending items")
	return tx.Commit()
}

func (s *adminService) GetSystemHealth() (*SystemHealth, error) {
	health := &SystemHealth{
		Status:   "ok",
		DBStatus: "Connected",
		Checks:   make(map[string]HealthCheck),
	}

	// Disk Space
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		free := stat.Bfree * uint64(stat.Bsize)
		total := stat.Blocks * uint64(stat.Bsize)
		health.DiskFree = formatBytes(free)
		health.DiskTotal = formatBytes(total)
		if total > 0 {
			usage := float64(total-free) / float64(total) * 100
			health.DiskUsagePct = usage

			status := "ok"
			msg := fmt.Sprintf("%s free of %s", health.DiskFree, health.DiskTotal)
			if usage > 90 {
				status = "critical"
				health.Status = "critical"
			} else if usage > 75 {
				status = "warning"
				if health.Status == "ok" {
					health.Status = "degraded"
				}
			}
			health.Checks["disk"] = HealthCheck{Status: status, Message: msg}
		}
	} else {
		health.DiskFree = "N/A"
		health.DiskTotal = "N/A"
		health.Checks["disk"] = HealthCheck{Status: "warning", Message: "Unable to read disk stats"}
	}

	// Memory Usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	health.MemoryUsed = formatBytes(m.Alloc)
	health.MemoryTotal = formatBytes(m.Sys)
	if m.Sys > 0 {
		usage := float64(m.Alloc) / float64(m.Sys) * 100
		health.MemoryUsagePct = usage

		status := "ok"
		msg := fmt.Sprintf("%s used of %s", health.MemoryUsed, health.MemoryTotal)
		if usage > 90 {
			status = "critical"
			health.Status = "critical"
		} else if usage > 80 {
			status = "warning"
			if health.Status == "ok" {
				health.Status = "degraded"
			}
		}
		health.Checks["memory"] = HealthCheck{Status: status, Message: msg}
	}

	// Check DB
	dbStatus := "ok"
	dbMsg := "Database is responsive"
	if err := s.db.Ping(); err != nil {
		health.DBStatus = "Disconnected"
		health.Status = "critical"
		dbStatus = "critical"
		dbMsg = "Database connection failed"
	}
	health.Checks["database"] = HealthCheck{Status: dbStatus, Message: dbMsg}

	// Uptime
	duration := time.Since(s.startTime)
	health.Uptime = formatDuration(duration)

	// Final status message
	switch health.Status {
	case "ok":
		health.Message = "All systems are operational."
	case "degraded":
		health.Message = "Some systems are experiencing issues."
	case "critical":
		health.Message = "Critical system failure detected."
	}

	return health, nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func (s *adminService) CreateBackup() (string, error) {
	backupDir := "/tmp/backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("backup_%s.sql", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(backupDir, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Manual dump logic (Simplified)
	tables := []string{"users", "products", "product_batches", "stock_entries", "sales", "sale_items", "activity_logs", "settings"}

	f.WriteString("-- Inventory System Backup\n")
	f.WriteString(fmt.Sprintf("-- Date: %s\n\n", time.Now().Format(time.RFC3339)))

	for _, table := range tables {
		f.WriteString(fmt.Sprintf("-- Table: %s\n", table))

		// SHOW CREATE TABLE
		var createTable sql.NullString
		err := s.db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s", table)).Scan(&table, &createTable)
		if err == nil && createTable.Valid {
			f.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", table))
			f.WriteString(createTable.String + ";\n\n")
		}

		// SELECT *
		rows, err := s.db.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			continue
		}

		cols, _ := rows.Columns()
		for rows.Next() {
			values := make([]interface{}, len(cols))
			valuePtrs := make([]interface{}, len(cols))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err == nil {
				var valStr []string
				for _, v := range values {
					if v == nil {
						valStr = append(valStr, "NULL")
					} else {
						switch t := v.(type) {
						case []byte:
							valStr = append(valStr, fmt.Sprintf("'%s'", strings.ReplaceAll(string(t), "'", "''")))
						case string:
							valStr = append(valStr, fmt.Sprintf("'%s'", strings.ReplaceAll(t, "'", "''")))
						default:
							valStr = append(valStr, fmt.Sprintf("%v", t))
						}
					}
				}
				f.WriteString(fmt.Sprintf("INSERT INTO `%s` VALUES (%s);\n", table, strings.Join(valStr, ", ")))
			}
		}
		rows.Close()
		f.WriteString("\n")
	}

	return filePath, nil
}

func (s *adminService) GetPendingItems() ([]map[string]interface{}, error) {
	return s.productRepo.GetPendingGrouped()
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("% d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

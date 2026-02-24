package service

import (
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"testing"
	"time"
)

type mockReportRepo struct {
	repository.ReportRepository
	lastToday string
}

func (m *mockReportRepo) GetTodaySummary(today string) (*repository.TodaySummary, error) {
	m.lastToday = today
	return &repository.TodaySummary{}, nil
}

func (m *mockReportRepo) GetYesterdaySummary(yesterday string) (*repository.TodaySummary, error) {
	return &repository.TodaySummary{}, nil
}

func (m *mockReportRepo) GetExpiringCount(days int, now string) (int, error) {
	return 0, nil
}

func (m *mockReportRepo) GetExpiredCount(now string) (int, error) {
	return 0, nil
}

func (m *mockReportRepo) GetWeeklySales(dates []string) ([]string, []float64, error) {
	return nil, nil, nil
}

func (m *mockReportRepo) GetProfitSummary(startDate, endDate string) (float64, error) {
	return 0, nil
}

type mockProductRepo struct {
	repository.ProductRepository
}

func (m *mockProductRepo) GetLowStockCount() (int, error)      { return 0, nil }
func (m *mockProductRepo) GetPendingPricesCount() (int, error) { return 0, nil }

type mockLogRepo struct {
	repository.ActivityLogRepository
}

func (m *mockLogRepo) Log(userID int, action string) error               { return nil }
func (m *mockLogRepo) GetLatest(limit int) ([]domain.ActivityLog, error) { return nil, nil }
func (m *mockLogRepo) Search(startDate, endDate, sort string) ([]domain.ActivityLog, error) {
	return nil, nil
}

type mockSettingService struct {
	SettingService
	tz string
}

func (m *mockSettingService) GetSettings() (*domain.Setting, error) {
	return &domain.Setting{Timezone: m.tz}, nil
}

func TestReportService_TimezoneLogic(t *testing.T) {
	reportRepo := &mockReportRepo{}
	productRepo := &mockProductRepo{}
	logRepo := &mockLogRepo{}

	// Test Asia/Makassar (UTC+8)
	// If it's early morning UTC (e.g. 23:00 on Feb 20), it's Feb 21 in Makassar.
	makassarService := NewReportService(reportRepo, productRepo, logRepo, &mockSettingService{tz: "Asia/Makassar"})

	_, err := makassarService.GetReportHubData()
	if err != nil {
		t.Fatalf("Failed to get hub data: %v", err)
	}

	now := time.Now().In(time.FixedZone("Makassar", 8*3600))
	expectedToday := now.Format("2006-01-02")

	if reportRepo.lastToday != expectedToday {
		t.Errorf("Expected today to be %s, got %s", expectedToday, reportRepo.lastToday)
	}
}

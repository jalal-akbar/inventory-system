package service

import (
	"encoding/csv"
	"fmt"
	"inventory-system/internal/repository"
	"io"
	"strconv"
	"time"
)

type ReportService interface {
	GetReportHubData() (map[string]interface{}, error)
	GetFinancialReportData(startDate, endDate, category string) (map[string]interface{}, error)
	GetPsychotropicReport(startDate, endDate string) ([]repository.PsychotropicReportRow, error)
	GetStockMutation(startDate, endDate string) ([]repository.StockMutationRow, error)
	ExportRecentSalesCSV(w io.Writer, startDate, endDate, category string) error
	ExportStockMutationCSV(w io.Writer, startDate, endDate string) error
	GetHistory(startDate, endDate string) ([]map[string]interface{}, error)
}

type reportService struct {
	reportRepo     repository.ReportRepository
	productRepo    repository.ProductRepository
	logRepo        repository.ActivityLogRepository
	settingService SettingService
}

func NewReportService(rRepo repository.ReportRepository, pRepo repository.ProductRepository, lRepo repository.ActivityLogRepository, sService SettingService) ReportService {
	return &reportService{
		reportRepo:     rRepo,
		productRepo:    pRepo,
		logRepo:        lRepo,
		settingService: sService,
	}
}

func (s *reportService) GetReportHubData() (map[string]interface{}, error) {
	// Get timezone from settings
	tz := "UTC"
	settings, err := s.settingService.GetSettings()
	if err == nil && settings != nil && settings.Timezone != "" {
		tz = settings.Timezone
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")
	yesterdayStr := now.AddDate(0, 0, -1).Format("2006-01-02")

	today, err := s.reportRepo.GetTodaySummary(todayStr)
	if err != nil {
		return nil, err
	}

	yesterday, err := s.reportRepo.GetYesterdaySummary(yesterdayStr)
	if err != nil {
		yesterday = &repository.TodaySummary{} // Fallback
	}

	revenueDelta := today.Revenue - yesterday.Revenue
	transactionsDelta := today.Transactions - yesterday.Transactions

	lowStock, _ := s.productRepo.GetLowStockCount()
	expiring, _ := s.reportRepo.GetExpiringCount(180, todayStr)
	expired, _ := s.reportRepo.GetExpiredCount(todayStr)
	pendingPrices, _ := s.productRepo.GetPendingPricesCount()

	todayProfit, _ := s.reportRepo.GetProfitSummary(todayStr, todayStr)

	// Weekly dates
	var weeklyDates []string
	for i := 6; i >= 0; i-- {
		weeklyDates = append(weeklyDates, now.AddDate(0, 0, -i).Format("2006-01-02"))
	}
	labels, values, _ := s.reportRepo.GetWeeklySales(weeklyDates)

	recentLogs, _ := s.logRepo.GetLatest(5)

	return map[string]interface{}{
		"todayRevenue":       today.Revenue,
		"todaySales":         today.Transactions,
		"revenueDelta":       revenueDelta,
		"salesDelta":         transactionsDelta,
		"lowStockCount":      lowStock,
		"expiringCount":      expiring,
		"expiredCount":       expired,
		"pendingPricesCount": pendingPrices,
		"todayProfit":        todayProfit,
		"todayPsychotropic":  today.PsychotropicSales,
		"weeklySales": map[string]interface{}{
			"labels": labels,
			"values": values,
		},
		"recentActivities": recentLogs,
	}, nil
}

func (s *reportService) GetFinancialReportData(startDate, endDate, category string) (map[string]interface{}, error) {
	summary, err := s.reportRepo.GetFinancialSummary(startDate, endDate, category)
	if err != nil {
		return nil, err
	}

	paymentBreakdown, err := s.reportRepo.GetPaymentMethodBreakdown(startDate, endDate, category)
	if err != nil {
		return nil, err
	}

	staffBreakdown, err := s.reportRepo.GetStaffBreakdown(startDate, endDate, category)
	if err != nil {
		return nil, err
	}

	recentSales, err := s.reportRepo.GetRecentSales(startDate, endDate, category, 50)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"summary":          summary,
		"paymentBreakdown": paymentBreakdown,
		"staffBreakdown":   staffBreakdown,
		"recentSales":      recentSales,
	}, nil
}

func (s *reportService) GetPsychotropicReport(startDate, endDate string) ([]repository.PsychotropicReportRow, error) {
	return s.reportRepo.GetPsychotropicReport(startDate, endDate)
}

func (s *reportService) GetStockMutation(startDate, endDate string) ([]repository.StockMutationRow, error) {
	return s.reportRepo.GetStockMutation(startDate, endDate)
}

func (s *reportService) GetHistory(startDate, endDate string) ([]map[string]interface{}, error) {
	return s.reportRepo.GetHistory(startDate, endDate)
}

func (s *reportService) ExportRecentSalesCSV(w io.Writer, startDate, endDate, category string) error {
	sales, err := s.reportRepo.GetRecentSales(startDate, endDate, category, 1000000)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"ID", "Date", "Staff", "Customer", "Payment Method", "Total Amount", "Profit", "Status"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, sale := range sales {
		row := []string{
			fmt.Sprintf("%v", sale["id"]),
			fmt.Sprintf("%v", sale["created_at"]),
			fmt.Sprintf("%v", sale["username"]),
			fmt.Sprintf("%v", sale["customer_name"]),
			fmt.Sprintf("%v", sale["payment_method"]),
			fmt.Sprintf("%v", sale["total_amount"]),
			fmt.Sprintf("%v", sale["profit"]),
			fmt.Sprintf("%v", sale["status"]),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func (s *reportService) ExportStockMutationCSV(w io.Writer, startDate, endDate string) error {
	reports, err := s.reportRepo.GetStockMutation(startDate, endDate)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"No", "Nama Obat", "Konversi", "Stok Awal (PCS)", "Masuk (PCS)", "Keluar (PCS)", "Stok Akhir (PCS)", "Satuan"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for i, r := range reports {
		stokAkhir := r.StokAwal + r.Masuk - r.Keluar
		konversi := "1"
		if r.ItemsPerUnit > 1 {
			konversi = fmt.Sprintf("1 %s = %d Pcs", r.Unit, r.ItemsPerUnit)
		}

		row := []string{
			strconv.Itoa(i + 1),
			r.Name,
			konversi,
			strconv.Itoa(r.StokAwal),
			strconv.Itoa(r.Masuk),
			strconv.Itoa(r.Keluar),
			strconv.Itoa(stokAkhir),
			r.Unit,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

package main

import (
	"context"
	"inventory-system/internal/config"
	"inventory-system/internal/handler"
	"inventory-system/internal/middleware"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting Inventory System")
	// 1. Database
	db, err := config.NewDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Initialize schema if needed
	if err := config.InitDB(db); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// 2. Repositories
	userRepo := repository.NewUserRepository(db)
	settingRepo := repository.NewSettingRepository(db)
	logRepo := repository.NewActivityLogRepository(db)
	pendingRepo := repository.NewPendingCountsRepository(db)
	productRepo := repository.NewProductRepository(db)
	batchRepo := repository.NewBatchRepository(db)
	entryRepo := repository.NewStockEntryRepository(db)
	saleRepo := repository.NewSaleRepository(db)
	reportRepo := repository.NewReportRepository(db)
	returnRepo := repository.NewReturnRepository(db)

	// 3. Services
	authService := service.NewAuthService(userRepo, logRepo)
	userService := service.NewUserService(userRepo, logRepo)
	settingService := service.NewSettingService(settingRepo, logRepo)
	productService := service.NewProductService(db, productRepo, batchRepo, entryRepo, logRepo)
	batchService := service.NewBatchService(db, batchRepo, entryRepo, logRepo)
	saleService := service.NewSaleService(db, saleRepo, productRepo, batchRepo, logRepo)
	adminService := service.NewAdminService(db, productRepo, batchRepo, entryRepo, saleRepo, logRepo)
	reportService := service.NewReportService(reportRepo, productRepo, logRepo, settingService)
	returnService := service.NewReturnService(db, returnRepo, saleRepo, batchRepo, logRepo)

	// 4. Handlers
	base := handler.BaseHandler{
		SettingService: settingService,
		PendingRepo:    pendingRepo,
	}
	authHandler := handler.NewAuthHandler(base, authService)
	userHandler := handler.NewUserHandler(base, userService, settingService, logRepo)
	productHandler := handler.NewProductHandler(base, productService, batchService)
	productApiHandler := handler.NewProductApiHandler(base, productService)
	batchApiHandler := handler.NewBatchApiHandler(base, batchService)
	inventoryApiHandler := handler.NewInventoryApiHandler(base, batchService)
	saleApiHandler := handler.NewSaleApiHandler(base, saleService)
	posHandler := handler.NewPosHandler(base, productService, saleService)
	adminHandler := handler.NewAdminHandler(base, adminService)
	reportHandler := handler.NewReportHandler(base, reportService)
	logHandler := handler.NewLogHandler(base, logRepo)
	dashboardHandler := handler.NewDashboardHandler(base, reportService, adminService, productService, batchService)
	returnHandler := handler.NewReturnHandler(base, returnService, saleService)

	// 5. Routes
	mux := http.NewServeMux()

	// Static Files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Auth
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)

	// Protected Page Routes
	mux.Handle("/dashboard", middleware.RequireAuth(http.HandlerFunc(dashboardHandler.Index)))
	mux.Handle("/reports", middleware.RequireAuth(http.HandlerFunc(reportHandler.Hub)))
	mux.Handle("/reports/financial", middleware.RequireAuth(http.HandlerFunc(reportHandler.Financial)))
	mux.Handle("/reports/history", middleware.RequireAuth(http.HandlerFunc(reportHandler.History)))
	mux.Handle("/reports/psychotropic", middleware.RequireAuth(http.HandlerFunc(reportHandler.Psychotropic)))
	mux.Handle("/reports/stock-mutation", middleware.RequireAuth(http.HandlerFunc(reportHandler.StockMutation)))
	mux.Handle("/admin/approvals", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(adminHandler.Approvals))))
	mux.Handle("/admin/backup", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(adminHandler.Backup))))
	mux.Handle("/activity-log", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(logHandler.Index))))

	mux.Handle("/products", middleware.RequireAuth(http.HandlerFunc(productHandler.Index)))
	mux.Handle("/products/check", middleware.RequireAuth(http.HandlerFunc(productHandler.InventoryCheck)))
	mux.Handle("/print-label", middleware.RequireAuth(http.HandlerFunc(productHandler.PrintLabel)))
	mux.Handle("/settings", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(userHandler.Settings))))
	mux.Handle("/pos", middleware.RequireAuth(http.HandlerFunc(posHandler.Index)))
	mux.Handle("/pos/search", middleware.RequireAuth(http.HandlerFunc(posHandler.Search)))
	mux.Handle("/pos/checkout", middleware.RequireAuth(http.HandlerFunc(posHandler.Checkout)))
	mux.Handle("/sales-history", middleware.RequireAuth(http.HandlerFunc(reportHandler.History)))
	mux.Handle("/sales/return", middleware.RequireAuth(http.HandlerFunc(returnHandler.ShowForm)))

	// API Routes
	// Products API
	mux.Handle("/api/products", middleware.RequireAuth(http.HandlerFunc(productApiHandler.Index)))
	mux.Handle("/api/products/create", middleware.RequireAuth(http.HandlerFunc(productApiHandler.Store)))
	mux.Handle("/api/products/update", middleware.RequireAuth(http.HandlerFunc(productApiHandler.Update)))
	mux.Handle("/api/products/delete", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(productApiHandler.Delete))))
	mux.Handle("/api/products/verify", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(productApiHandler.Verify))))
	mux.Handle("/api/products/details", middleware.RequireAuth(http.HandlerFunc(productApiHandler.GetDetails)))
	mux.Handle("/api/products/pending-count", middleware.RequireAuth(http.HandlerFunc(productApiHandler.GetPendingCount)))
	mux.Handle("/api/products/search", middleware.RequireAuth(http.HandlerFunc(productApiHandler.Search)))

	// Batches API
	mux.Handle("/api/batches", middleware.RequireAuth(http.HandlerFunc(batchApiHandler.Index)))
	mux.Handle("/api/batches/create", middleware.RequireAuth(http.HandlerFunc(batchApiHandler.Store)))

	// Inventory API
	mux.Handle("/api/inventory/adjust", middleware.RequireAuth(http.HandlerFunc(inventoryApiHandler.Adjust)))

	// Sales API
	mux.Handle("/api/sales/create", middleware.RequireAuth(http.HandlerFunc(saleApiHandler.Store)))
	mux.Handle("/api/sales/void", middleware.RequireAuth(http.HandlerFunc(saleApiHandler.Void)))
	mux.Handle("/api/sales/details", middleware.RequireAuth(http.HandlerFunc(saleApiHandler.GetDetails)))
	mux.Handle("/api/sales/details-html", middleware.RequireAuth(http.HandlerFunc(saleApiHandler.GetDetailsHTML)))

	// Return API
	mux.Handle("/api/returns/create", middleware.RequireAuth(http.HandlerFunc(returnHandler.Store)))

	// Admin API
	mux.Handle("/api/admin/pending", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(adminHandler.GetPendingItems))))
	mux.Handle("/api/admin/approve", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(adminHandler.ApproveItem))))
	mux.Handle("/api/admin/reject", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(adminHandler.RejectItem))))
	mux.Handle("/api/admin/approve-group", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(adminHandler.ApproveGroup))))
	mux.Handle("/api/admin/approve-all", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(adminHandler.ApproveAll))))
	mux.Handle("/admin/backup/create", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(adminHandler.CreateBackup))))

	// Dashboard API
	mux.Handle("/api/dashboard/health", middleware.RequireAuth(http.HandlerFunc(dashboardHandler.Health)))
	mux.Handle("/api/dashboard/hub-data", middleware.RequireAuth(http.HandlerFunc(dashboardHandler.GetHubData)))

	// User API
	mux.Handle("/api/user/change-password", middleware.RequireAuth(http.HandlerFunc(userHandler.ChangePassword)))
	mux.Handle("/api/user/set-lang", middleware.RequireAuth(http.HandlerFunc(userHandler.SetLanguage)))
	mux.Handle("/api/user/update-username", middleware.RequireAuth(http.HandlerFunc(userHandler.UpdateUsername)))
	mux.Handle("/api/users/create", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(userHandler.CreateUser))))
	mux.Handle("/api/users/toggle-status", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(userHandler.ToggleUserStatus))))
	mux.Handle("/api/settings/update-global", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(userHandler.UpdateSettings))))

	// Report Export API
	mux.Handle("/reports/export-csv", middleware.RequireAuth(http.HandlerFunc(reportHandler.ExportCSV)))
	mux.Handle("/reports/export-mutation-csv", middleware.RequireAuth(http.HandlerFunc(reportHandler.ExportStockMutationCSV)))

	// Main Redirect
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// 6. Middleware Chain
	app := middleware.Logger(middleware.SessionMiddleware(middleware.CSRFProtect(mux)))

	// 7. Configure Server
	srv := &http.Server{
		Addr:         ":7070",
		Handler:      app,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Start Server
	go func() {
		log.Println("Server starting on :7070...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 9. Wait for Shutdown Signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Closing database connection...")
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server exited gracefully")
}

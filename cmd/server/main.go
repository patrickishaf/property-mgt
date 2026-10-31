package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/cache"
	"github.com/patrickishaf/property-mgt/internal/db"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"github.com/patrickishaf/property-mgt/internal/net"
	"github.com/patrickishaf/property-mgt/pkg/apidocs"
	"github.com/patrickishaf/property-mgt/pkg/auth"
	"github.com/patrickishaf/property-mgt/pkg/maintenance"
	"github.com/patrickishaf/property-mgt/pkg/messaging"
	"github.com/patrickishaf/property-mgt/pkg/payment"
	"github.com/patrickishaf/property-mgt/pkg/property"
	"github.com/patrickishaf/property-mgt/pkg/settings"
	"github.com/patrickishaf/property-mgt/pkg/tenant"
	"github.com/patrickishaf/property-mgt/providers/cpaas"
)

func main() {
	cfg := config.LoadConfig()
	logger := common.CreateLogger()
	dbConn := db.Connect(&cfg.DB)
	cacheClient := cache.CreateClient(&cfg.Redis)
	scheduler := common.CreateScheduler()

	// cpaas
	cpaasClient := cpaas.NewClient(&cfg.Cpaas)
	cpaas.NewWorker(cpaasClient, scheduler)

	// repositories
	adminRepo := repository.CreateAdminRepository(dbConn, logger)
	jobRepo := repository.CreateJobRepository(&cfg.App, dbConn, logger)
	leaseRepo := repository.CreateLeaseRepo(dbConn, logger)
	propertyRepo := repository.CreatePropertyRepository(dbConn, logger)
	tenantRepo := repository.CreateTenantRepository(dbConn, logger)
	unitRepo := repository.CreateUnitRepository(dbConn, logger)

	// services
	authService := auth.CreateService(&cfg.App, logger, adminRepo, tenantRepo)
	maintenanceService := maintenance.CreateService()
	messagingService := messaging.CreateService()
	paymentService := payment.CreateService()
	propertyService := property.CreateService(&cfg.App, logger, dbConn, propertyRepo, unitRepo, leaseRepo, adminRepo, jobRepo)
	settingsService := settings.CreateService()
	tenantService := tenant.CreateService(&cfg.App, logger, dbConn, tenantRepo, leaseRepo, propertyRepo, jobRepo, unitRepo)

	// http handlers
	authHTTPHandler := auth.CreateHTTPHandler(authService)
	maintenanceHTTPHandler := maintenance.CreateHTTPHandler(maintenanceService)
	messagingHTTPHandler := messaging.CreateHTTPHandler(messagingService)
	paymentHTTPHandler := payment.CreateHTTPHandler(paymentService)
	propertyHTTPHandler := property.CreateHTTPHandler(propertyService)
	settingsHTTPHandler := settings.CreateHTTPHandler(settingsService)
	tenantHTTPHandler := tenant.CreateHTTPHandler(tenantService)

	// routers
	authRouter := auth.CreateRouter(&cfg.App, authHTTPHandler, "/auth")
	maintenanceRouter := maintenance.CreateRouter(maintenanceHTTPHandler, "/maintenance")
	messagingRouter := messaging.CreateRouter(messagingHTTPHandler, "/messages")
	paymentRouter := payment.CreateRouter(paymentHTTPHandler, "/payments")
	propertyRouter := property.CreateRouter(&cfg.App, propertyHTTPHandler, "/properties")
	settingsRouter := settings.CreateRouter(settingsHTTPHandler, "/settings")
	tenantRouter := tenant.CreateRouter(&cfg.App, tenantHTTPHandler, "/tenants")

	srv := net.CreateServer(&cfg.App)
	srv.ConfigureMidleware([]net.Middleware{
		net.AttachRequestID,
	})
	srv.ConfigureRoutes("", []common.AppRouter{
		authRouter,
		maintenanceRouter,
		messagingRouter,
		paymentRouter,
		propertyRouter,
		settingsRouter,
		tenantRouter,
	})
	if cfg.App.Env == "development" {
		srv.ConfigureRoutes("", []common.AppRouter{apidocs.CreateRouter("/docs")})
	}
	go srv.Run()
	scheduler.Start()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-exitChan
	logger.Debug("interrupt signal received. shutting down system")
	scheduler.Stop()
	cacheClient.Client().Close()
	db.CloseDB()
	os.Exit(0)
}

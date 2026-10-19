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
	connPool := db.Connect(&cfg.DB)
	cacheClient := cache.CreateClient(&cfg.Redis)
	scheduler := common.CreateScheduler()

	// cpaas
	cpaasClient := cpaas.NewClient(&cfg.Cpaas)
	cpaas.NewWorker(cpaasClient, scheduler)

	// repositories
	repository.CreateJobRepository(&cfg.App, connPool, logger)

	// services
	authService := auth.CreateService()
	maintenanceService := maintenance.CreateService()
	messagingService := messaging.CreateService()
	paymentService := payment.CreateService()
	propertyService := property.CreateService()
	settingsService := settings.CreateService()
	tenantService := tenant.CreateService()

	// http handlers
	authHTTPHandler := auth.CreateHTTPHandler(authService)
	maintenanceHTTPHandler := maintenance.CreateHTTPHandler(maintenanceService)
	messagingHTTPHandler := messaging.CreateHTTPHandler(messagingService)
	paymentHTTPHandler := payment.CreateHTTPHandler(paymentService)
	propertyHTTPHandler := property.CreateHTTPHandler(propertyService)
	settingsHTTPHandler := settings.CreateHTTPHandler(settingsService)
	tenantHTTPHandler := tenant.CreateHTTPHandler(tenantService)

	// routers
	authRouter := auth.CreateRouter(authHTTPHandler, "/auth")
	maintenanceRouter := maintenance.CreateRouter(maintenanceHTTPHandler, "/maintenance")
	messagingRouter := messaging.CreateRouter(messagingHTTPHandler, "/messages")
	paymentRouter := payment.CreateRouter(paymentHTTPHandler, "/payments")
	propertyRouter := property.CreateRouter(propertyHTTPHandler, "/properties")
	settingsRouter := settings.CreateRouter(settingsHTTPHandler, "/settings")
	tenantRouter := tenant.CreateRouter(tenantHTTPHandler, "/tenants")

	srv := net.CreateServer(&cfg.App)
	srv.ConfigureMidleware([]net.Middleware{})
	srv.ConfigureRoutes("", []common.AppRouter{
		authRouter,
		maintenanceRouter,
		messagingRouter,
		paymentRouter,
		propertyRouter,
		settingsRouter,
		tenantRouter,
	})
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

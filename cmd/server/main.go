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
)

func main() {
	cfg := config.LoadConfig()
	logger := common.CreateLogger()
	connPool := db.Connect(&cfg.DB)
	cacheClient := cache.CreateClient(&cfg.Redis)

	// repositories
	repository.CreateJobRepository(&cfg.App, connPool, logger)

	// services
	authService := auth.CreateService()

	// http handlers
	authHTTPHandler := auth.CreateHTTPHandler(authService)

	// routers
	authRouter := auth.CreateRouter(authHTTPHandler, "/auth")

	srv := net.CreateServer(&cfg.App)
	srv.ConfigureMidleware([]net.Middleware{})
	srv.ConfigureRoutes("", []common.AppRouter{authRouter})
	go srv.Run()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-exitChan
	logger.Debug("interrupt signal received. shutting down system")
	cacheClient.Client().Close()
	db.CloseDB()
	os.Exit(0)
}

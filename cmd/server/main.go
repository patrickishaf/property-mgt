package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/cache"
	"github.com/patrickishaf/property-mgt/internal/db"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"github.com/patrickishaf/property-mgt/internal/net"
)

func main() {
	cfg := config.LoadConfig()
	logger := common.CreateLogger()
	connPool, err := db.CreateConnectionPool(context.Background(), cfg.DB.GetConnectionString())
	if err != nil {
		logger.Error("failed to create db connection pool", "err", err.Error())
		panic(err)
	}
	cacheClient := cache.CreateClient(&cfg.Redis)

	// repositories
	repository.CreateJobRepository(&cfg.App, connPool, logger)
	srv := net.CreateServer(&cfg.App)
	srv.ConfigureMidleware([]net.Middleware{})
	go srv.Run()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-exitChan
	logger.Debug("interrupt signal received. shutting down system")
	cacheClient.Client().Close()
	connPool.Close()
	os.Exit(0)
}

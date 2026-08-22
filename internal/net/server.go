package net

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
)

type Server struct {
	cfg     *config.AppConfig
	handler *gin.Engine
}

func CreateServer(cfg *config.AppConfig) *Server {
	return &Server{
		cfg:     cfg,
		handler: gin.Default(),
	}
}

func (this *Server) ConfigureRoutes(pathPrefix string, routers []common.AppRouter) {
	for _, router := range routers {
		group := this.handler.Group(fmt.Sprintf("%s%s", pathPrefix, router.GetPath()))
		router.RegisterRoutes(group)
	}
}

func (this *Server) GetAddr() string {
	return fmt.Sprintf(":%d", this.cfg.Port)
}

func (this *Server) ConfigureMidleware(middlewares []Middleware) {
	this.handler.Use(gin.Recovery())
	this.handler.Use(cors.Default())
	for _, mdlwr := range middlewares {
		this.handler.Use(mdlwr())
	}
}

func (this *Server) Run() error {
	return this.handler.Run(this.GetAddr())
}

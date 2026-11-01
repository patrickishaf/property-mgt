package net

import (
	"fmt"
	"log/slog"
	"strings"

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
	this.handler.Use(cors.New(corsConfig(this.cfg)))
	for _, mdlwr := range middlewares {
		this.handler.Use(mdlwr())
	}
}

func (this *Server) Run() error {
	return this.handler.Run(this.GetAddr())
}

// corsConfig allows the origins in ALLOWED_ORIGINS ("*" allows any). With none
// set, development allows any origin and other environments allow none, so a
// missing setting never opens production up.
func corsConfig(cfg *config.AppConfig) cors.Config {
	corsCfg := cors.DefaultConfig()
	corsCfg.AddAllowHeaders("Authorization")
	var origins []string
	for _, origin := range strings.Split(cfg.AllowedOrigins, ",") {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin == "*" {
			corsCfg.AllowAllOrigins = true
			return corsCfg
		}
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		if cfg.Env == "development" {
			corsCfg.AllowAllOrigins = true
			return corsCfg
		}
		slog.Warn("ALLOWED_ORIGINS is empty: browsers on other origins can't call the API")
		corsCfg.AllowOriginFunc = func(string) bool { return false }
		return corsCfg
	}
	corsCfg.AllowOrigins = origins
	return corsCfg
}

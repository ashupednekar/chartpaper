package server

import (
	"log"
)

func (s *Server) buildRoutes() error{
  chartpaper := s.engine.Group("/chartpaper")
	api := chartpaper.Group("/api")
	{
		api.GET("/health", s.healthCheck)
		api.GET("/charts", s.getStoredCharts)
		api.GET("/charts/:name", s.getStoredChartInfo)
		api.GET("/charts/:name/versions", s.getChartVersions)
		api.GET("/charts/:name/dependencies", s.getChartDependencies)
		api.POST("/charts/:name/fetch-dependencies", s.fetchChartDependencies)
		api.POST("/charts/:name/switch-version", s.switchChartVersion)
		api.GET("/docker-config", s.getDockerConfig)
		api.POST("/fetch-chart", s.fetchChart)
		api.POST("/authenticate", s.authenticate)
		api.DELETE("/charts/:name", s.deleteChart)
		api.DELETE("/charts/:name/versions/:version", s.deleteChartVersion)
		api.GET("/registry-configs", s.getRegistryConfigs)
		api.POST("/registry-configs", s.createRegistryConfig)
		api.PUT("/registry-configs/:id", s.updateRegistryConfig)
		api.DELETE("/registry-configs/:id", s.deleteRegistryConfig)
		api.POST("/registry-configs/:id/set-default", s.setDefaultRegistry)
	}
	
	log.Printf("=== API ROUTES REGISTERED ===\n")
	log.Printf("GET /api/health\n")
	log.Printf("GET /api/charts (DATABASE ONLY - NO DIRECTORY SCANNING)\n")
	log.Printf("GET /api/charts/:name\n")
	log.Printf("GET /api/charts/:name/dependencies\n")
	log.Printf("GET /api/docker-config\n")
	log.Printf("POST /api/fetch-chart\n")
	log.Printf("POST /api/authenticate\n")
	log.Printf("DELETE /api/charts/:name\n")

	log.Printf("=== SERVER STARTING ON :8000 ===\n")
	log.Printf("NO DIRECTORY SCANNING - OCI REGISTRY ONLY\n")
	return nil
}

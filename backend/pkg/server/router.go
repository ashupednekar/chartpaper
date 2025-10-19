package server

func (s *Server) buildRoutes() error{
	s.engine.GET("/healthz", s.healthCheck)
	s.engine.GET("/livez", s.livenessCheck)
  chartpaper := s.engine.Group("/chartpaper")
	api := chartpaper.Group("/api")
	{
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
	return nil
}

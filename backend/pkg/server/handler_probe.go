package server

import (
	"chartpaper/internal/db"
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)


func (s *Server) livenessCheck(c *gin.Context) {
}

func (s *Server) healthCheck(c *gin.Context) {
	queries := db.New(s.db)
	log.Printf("GET /api/health - Health check requested\n")
	
	if err := s.db.Ping(); err != nil {
		log.Printf("Database health check failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"database": "disconnected",
			"error": err.Error(),
		})
		return
	}
	
	ctx := context.Background()
	charts, err := queries.ListCharts(ctx)
	if err != nil {
		log.Printf("Database query failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"database": "connected",
			"query": "failed",
			"error": err.Error(),
		})
		return
	}
	
	log.Printf("Health check passed - found %d charts\n", len(charts))
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"database": "connected",
		"charts_count": len(charts),
		"version": "1.0.0",
	})
}



package server

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) buildMiddleWares() error{
	s.engine.Use(func(c *gin.Context) {
		log.Printf("\n🔥 INCOMING REQUEST 🔥\n")
		log.Printf("Method: %s\n", c.Request.Method)
		log.Printf("URL: %s\n", c.Request.URL.String())
		log.Printf("Path: %s\n", c.Request.URL.Path)
		log.Printf("User-Agent: %s\n", c.Request.Header.Get("User-Agent"))
		log.Printf("Origin: %s\n", c.Request.Header.Get("Origin"))
		log.Printf("========================\n")
		c.Next()
		log.Printf("Response Status: %d\n", c.Writer.Status())
		log.Printf("🔥 REQUEST COMPLETE 🔥\n\n")
	})
	
	s.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	return nil
}

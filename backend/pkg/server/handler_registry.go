package server

import (
	"chartpaper/pkg"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ashupednekar/compose/pkg/charts"
	"github.com/ashupednekar/compose/pkg/spec"
	"github.com/gin-gonic/gin"
)

func (s *Server) getDockerConfig(c *gin.Context) {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "compose", "config.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Docker config not found"})
		return
	}
	
	var config pkg.DockerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse docker config"})
		return
	}
	
	response := map[string]string{
		"username": config.Username,
		"registry": config.Registry,
		"status":   "configured",
	}
	
	c.JSON(http.StatusOK, response)
}

func (s *Server) authenticate(c *gin.Context) {
	
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "compose", "config.json")
	
	log.Printf("Looking for config at: %s\n", configPath)
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Config read error: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Docker config not found", 
			"path": configPath,
			"details": err.Error(),
		})
		return
	}
	
	log.Printf("Config data: %s\n", string(data))
	
	var config pkg.DockerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Config parse error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse docker config",
			"details": err.Error(),
		})
		return
	}
	
	log.Printf("Parsed config: Username=%s, Registry=%s\n", config.Username, config.Registry)
	
	chartUtils, err := charts.NewChartUtils()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize chart utils"})
		return
	}
	
	authInfo := &spec.AuthInfo{
		Username: config.Username,
		Password: config.Password,
		Registry: config.Registry,
	}
	
	if err := chartUtils.Authenticate(authInfo); err != nil {
		log.Printf("Authentication error: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": fmt.Sprintf("Authentication failed: %v", err),
			"registry": config.Registry,
			"username": config.Username,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"registry": config.Registry,
		"username": config.Username,
	})
}

func (s *Server) getRegistryConfigs(c *gin.Context) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, registry_url, username, password, is_default, created_at, updated_at 
		FROM registry_configs 
		ORDER BY is_default DESC, name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	
	var configs []pkg.RegistryConfig
	for rows.Next() {
		var config pkg.RegistryConfig
		var username, password sql.NullString
		err := rows.Scan(&config.ID, &config.Name, &config.RegistryURL, &username, 
			&password, &config.IsDefault, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		
		if username.Valid {
			config.Username = username.String
		}
		if password.Valid {
			config.Password = password.String
		}
		
		configs = append(configs, config)
	}
	
	c.JSON(http.StatusOK, configs)
}

func (s *Server) createRegistryConfig(c *gin.Context) {
	ctx := context.Background()
	
	var config pkg.RegistryConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	
	if config.IsDefault {
		_, err := s.db.ExecContext(ctx, "UPDATE registry_configs SET is_default = FALSE")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing defaults"})
			return
		}
	}
	
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO registry_configs (name, registry_url, username, password, is_default) 
		VALUES ($1, $2, $3, $4, $5)
	`, config.Name, config.RegistryURL, config.Username, config.Password, config.IsDefault)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	id, _ := result.LastInsertId()
	config.ID = id
	
	c.JSON(http.StatusCreated, config)
}


func (s *Server) updateRegistryConfig(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	
	var config pkg.RegistryConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	
	if config.IsDefault {
		_, err := s.db.ExecContext(ctx, "UPDATE registry_configs SET is_default = FALSE WHERE id != $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing defaults"})
			return
		}
	}
	
	_, err := s.db.ExecContext(ctx, `
		UPDATE registry_configs 
		SET name = $1, registry_url = $2, username = $3, password = $4, is_default = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`, config.Name, config.RegistryURL, config.Username, config.Password, config.IsDefault, id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, config)
}


func (s *Server) deleteRegistryConfig(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	
	_, err := s.db.ExecContext(ctx, "DELETE FROM registry_configs WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Registry configuration deleted"})
}


func (s *Server) setDefaultRegistry(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	
	
	_, err := s.db.ExecContext(ctx, "UPDATE registry_configs SET is_default = FALSE")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing defaults"})
		return
	}
	
	
	_, err = s.db.ExecContext(ctx, "UPDATE registry_configs SET is_default = TRUE WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Default registry updated"})
}

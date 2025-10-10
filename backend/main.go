package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"helm-visualizer/db"

	"github.com/ashupednekar/compose/pkg/charts"
	"github.com/ashupednekar/compose/pkg/spec"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"helm.sh/helm/v3/pkg/release"
)

type Chart struct {
	APIVersion   string       `yaml:"apiVersion" json:"apiVersion"`
	Name         string       `yaml:"name" json:"name"`
	Version      string       `yaml:"version" json:"version"`
	Description  string       `yaml:"description" json:"description"`
	Type         string       `yaml:"type" json:"type"`
	Dependencies []Dependency `yaml:"dependencies" json:"dependencies"`
}

type Dependency struct {
	Name       string `yaml:"name" json:"name"`
	Version    string `yaml:"version" json:"version"`
	Repository string `yaml:"repository" json:"repository"`
	Condition  string `yaml:"condition,omitempty" json:"condition,omitempty"`
}

type Values struct {
	Image  ImageConfig  `yaml:"image" json:"image"`
	Canary CanaryConfig `yaml:"canary" json:"canary"`
}

type ImageConfig struct {
	Tag string `yaml:"tag" json:"tag"`
}

type CanaryConfig struct {
	Tag string `yaml:"tag" json:"tag"`
}

type ChartInfo struct {
	Chart            Chart             `json:"chart"`
	ImageTag         string            `json:"imageTag"`
	CanaryTag        string            `json:"canaryTag"`
	ManifestMetadata *ManifestMetadata `json:"manifestMetadata,omitempty"`
}

type DockerConfig struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
	Registry string `json:"Registry"`
}

type ChartRequest struct {
	ChartURL    string   `json:"chartUrl"`
	ValuesPath  string   `json:"valuesPath"`
	SetValues   []string `json:"setValues"`
	UseHostNetwork bool  `json:"useHostNetwork"`
}

var database *sql.DB
var queries *db.Queries

func main() {
	fmt.Printf("=== CHART PAPER BACKEND STARTING ===\n")
	fmt.Printf("Version: OCI-only (no directory scanning)\n")
	
	// Initialize database
	var err error
	database, err = sql.Open("sqlite3", "./charts.db")
	if err != nil {
		panic(fmt.Sprintf("Failed to open database: %v", err))
	}
	defer database.Close()

	// Initialize database schema
	if err := initDB(); err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	queries = db.New(database)
	fmt.Printf("Database initialized and queries ready\n")

	r := gin.Default()
	
	// Custom logging middleware to track ALL requests
	r.Use(func(c *gin.Context) {
		fmt.Printf("\n🔥 INCOMING REQUEST 🔥\n")
		fmt.Printf("Method: %s\n", c.Request.Method)
		fmt.Printf("URL: %s\n", c.Request.URL.String())
		fmt.Printf("Path: %s\n", c.Request.URL.Path)
		fmt.Printf("User-Agent: %s\n", c.Request.Header.Get("User-Agent"))
		fmt.Printf("Origin: %s\n", c.Request.Header.Get("Origin"))
		fmt.Printf("========================\n")
		c.Next()
		fmt.Printf("Response Status: %d\n", c.Writer.Status())
		fmt.Printf("🔥 REQUEST COMPLETE 🔥\n\n")
	})
	
	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/health", healthCheck)
		api.GET("/charts", getStoredCharts)
		api.GET("/charts/:name", getStoredChartInfo)
		api.GET("/charts/:name/versions", getChartVersions)
		api.GET("/charts/:name/dependencies", getChartDependencies)
		api.POST("/charts/:name/fetch-dependencies", fetchChartDependenciesAdvanced)
		api.POST("/charts/:name/switch-version", switchChartVersion)
		api.GET("/docker-config", getDockerConfig)
		api.POST("/fetch-chart", fetchChart)
		api.POST("/authenticate", authenticate)
		api.DELETE("/charts/:name", deleteChart)
		api.DELETE("/charts/:name/versions/:version", deleteChartVersion)
		
		// Registry configuration endpoints
		api.GET("/registry-configs", getRegistryConfigs)
		api.POST("/registry-configs", createRegistryConfig)
		api.PUT("/registry-configs/:id", updateRegistryConfig)
		api.DELETE("/registry-configs/:id", deleteRegistryConfig)
		api.POST("/registry-configs/:id/set-default", setDefaultRegistry)
		
		api.GET("/test", func(c *gin.Context) {
			fmt.Printf("🚀 TEST ENDPOINT HIT - NEW CODE IS RUNNING!\n")
			c.JSON(http.StatusOK, gin.H{
				"message": "NEW CODE IS RUNNING - NO DIRECTORY SCANNING",
				"version": "2.0.0-oci-only",
				"timestamp": "2025-01-10",
			})
		})
	}
	
	fmt.Printf("=== API ROUTES REGISTERED ===\n")
	fmt.Printf("GET /api/health\n")
	fmt.Printf("GET /api/charts (DATABASE ONLY - NO DIRECTORY SCANNING)\n")
	fmt.Printf("GET /api/charts/:name\n")
	fmt.Printf("GET /api/charts/:name/dependencies\n")
	fmt.Printf("GET /api/docker-config\n")
	fmt.Printf("POST /api/fetch-chart\n")
	fmt.Printf("POST /api/authenticate\n")
	fmt.Printf("DELETE /api/charts/:name\n")

	fmt.Printf("=== SERVER STARTING ON :8000 ===\n")
	fmt.Printf("NO DIRECTORY SCANNING - OCI REGISTRY ONLY\n")
	r.Run(":8000")
}

func initDB() error {
	fmt.Printf("Initializing database...\n")
	
	// Check if tables already exist
	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='charts'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing tables: %v", err)
	}
	
	if count == 0 {
		// Tables don't exist, create them
		schema, err := os.ReadFile("db/schema/001_charts.sql")
		if err != nil {
			return fmt.Errorf("failed to read schema file: %v", err)
		}

		fmt.Printf("Executing database schema...\n")
		_, err = database.Exec(string(schema))
		if err != nil {
			return fmt.Errorf("failed to execute schema: %v", err)
		}
		fmt.Printf("Database schema created successfully\n")
	} else {
		fmt.Printf("Database tables already exist\n")
	}
	
	// Check if is_latest column exists (migration check)
	var columnExists int
	err = database.QueryRow("SELECT COUNT(*) FROM pragma_table_info('charts') WHERE name='is_latest'").Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check is_latest column: %v", err)
	}
	
	if columnExists == 0 {
		fmt.Printf("Running migration to add version history support...\n")
		migration, err := os.ReadFile("db/schema/002_add_version_history.sql")
		if err != nil {
			return fmt.Errorf("failed to read migration file: %v", err)
		}

		_, err = database.Exec(string(migration))
		if err != nil {
			return fmt.Errorf("failed to execute migration: %v", err)
		}
		fmt.Printf("Migration completed successfully\n")
	} else {
		fmt.Printf("Version history support already exists\n")
	}
	
	// Check if dependency tags columns exist (migration check)
	var depTagsExists int
	err = database.QueryRow("SELECT COUNT(*) FROM pragma_table_info('dependencies') WHERE name='image_tag'").Scan(&depTagsExists)
	if err != nil {
		return fmt.Errorf("failed to check dependency tags columns: %v", err)
	}
	
	if depTagsExists == 0 {
		fmt.Printf("Running migration to add dependency tags...\n")
		migration, err := os.ReadFile("db/schema/003_add_dependency_tags.sql")
		if err != nil {
			return fmt.Errorf("failed to read dependency tags migration file: %v", err)
		}

		_, err = database.Exec(string(migration))
		if err != nil {
			return fmt.Errorf("failed to execute dependency tags migration: %v", err)
		}
		fmt.Printf("Dependency tags migration completed successfully\n")
	} else {
		fmt.Printf("Dependency tags support already exists\n")
	}
	
	// Check if manifest metadata columns exist (migration check)
	var manifestMetadataExists int
	err = database.QueryRow("SELECT COUNT(*) FROM pragma_table_info('charts') WHERE name='ingress_paths'").Scan(&manifestMetadataExists)
	if err != nil {
		return fmt.Errorf("failed to check manifest metadata columns: %v", err)
	}
	
	if manifestMetadataExists == 0 {
		fmt.Printf("Running migration to add manifest metadata...\n")
		migration, err := os.ReadFile("db/schema/003_add_manifest_metadata.sql")
		if err != nil {
			return fmt.Errorf("failed to read manifest metadata migration file: %v", err)
		}

		_, err = database.Exec(string(migration))
		if err != nil {
			return fmt.Errorf("failed to execute manifest metadata migration: %v", err)
		}
		fmt.Printf("Manifest metadata migration completed successfully\n")
	} else {
		fmt.Printf("Manifest metadata support already exists\n")
	}

	fmt.Printf("Database initialized successfully\n")
	return nil
}

func storeChartInDB(chartInfo ChartInfo, apps []spec.App, chartURL string) (*db.Chart, error) {
	ctx := context.Background()
	
	// Simple approach: always try to get the existing chart first
	existingChart, err := queries.GetChart(ctx, chartInfo.Chart.Name)
	
	var storedChart db.Chart
	
	if err != nil {
		// Chart doesn't exist, create new one
		fmt.Printf("📝 Creating new chart: %s\n", chartInfo.Chart.Name)
		storedChart, err = queries.CreateChart(ctx, db.CreateChartParams{
			Name:        chartInfo.Chart.Name,
			Version:     chartInfo.Chart.Version,
			Description: sql.NullString{String: chartInfo.Chart.Description, Valid: chartInfo.Chart.Description != ""},
			Type:        chartInfo.Chart.Type,
			ChartUrl:    chartURL,
			ImageTag:    sql.NullString{String: chartInfo.ImageTag, Valid: chartInfo.ImageTag != "N/A"},
			CanaryTag:   sql.NullString{String: chartInfo.CanaryTag, Valid: chartInfo.CanaryTag != "N/A"},
			Manifest:    sql.NullString{String: "", Valid: false},
			IsLatest:    sql.NullBool{Bool: true, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create chart: %v", err)
		}
		
		// Update manifest metadata if available
		if chartInfo.ManifestMetadata != nil {
			err = updateChartManifestMetadata(storedChart.ID, *chartInfo.ManifestMetadata)
			if err != nil {
				fmt.Printf("⚠️  Warning: failed to update manifest metadata: %v\n", err)
			}
		}
	} else {
		// Chart exists, use the existing one
		fmt.Printf("📝 Using existing chart: %s (ID: %d)\n", chartInfo.Chart.Name, existingChart.ID)
		storedChart = existingChart
		
		// Clear existing dependencies
		queries.DeleteChartDependencies(ctx, storedChart.ID)
	}
	
	fmt.Printf("🚀 REACHED DEPENDENCY STORAGE SECTION!\n")
	fmt.Printf("🔍 DEBUG: storedChart pointer: %p\n", &storedChart)
	fmt.Printf("🔍 DEBUG: storedChart ID: %d\n", storedChart.ID)
	
	// Store dependencies
	fmt.Printf("🔍 DEBUG: About to store %d dependencies for chart %s (ID: %d)\n", len(chartInfo.Chart.Dependencies), chartInfo.Chart.Name, storedChart.ID)
	fmt.Printf("🔍 DEBUG: Dependencies array: %+v\n", chartInfo.Chart.Dependencies)
	for i, dep := range chartInfo.Chart.Dependencies {
		fmt.Printf("🔍 DEBUG: Processing dependency %d: %+v\n", i+1, dep)
		
		// Try to fetch dependency chart info to get image/canary tags
		var depImageTag, depCanaryTag string = "N/A", "N/A"
		
		if dep.Repository != "" {
			// Try to fetch the dependency chart to get its image tags
			depChartURL := dep.Repository
			if !strings.HasSuffix(depChartURL, "/"+dep.Name) {
				depChartURL = depChartURL + "/" + dep.Name
			}
			
			fmt.Printf("🔍 Attempting to fetch dependency info from: %s\n", depChartURL)
			if depChartInfo, depErr := tryFetchChart(depChartURL, dep.Name, dep.Version); depErr == nil {
				depImageTag = depChartInfo.ImageTag
				depCanaryTag = depChartInfo.CanaryTag
				fmt.Printf("✅ Got dependency tags: image=%s, canary=%s\n", depImageTag, depCanaryTag)
			} else {
				fmt.Printf("⚠️ Could not fetch dependency info: %v\n", depErr)
			}
		}
		
		// For now, store dependency without tags until we can update the schema properly
		depResult, err := queries.CreateDependency(ctx, db.CreateDependencyParams{
			ChartID:           storedChart.ID,
			DependencyName:    dep.Name,
			DependencyVersion: dep.Version,
			Repository:        sql.NullString{String: dep.Repository, Valid: dep.Repository != ""},
			ConditionField:    sql.NullString{String: dep.Condition, Valid: dep.Condition != ""},
		})
		if err != nil {
			fmt.Printf("❌ ERROR: failed to store dependency %s: %v\n", dep.Name, err)
		} else {
			fmt.Printf("✅ SUCCESS: Stored dependency: %s v%s (ID: %d)\n", dep.Name, dep.Version, depResult.ID)
		}
	}

	// Store apps
	for _, app := range apps {
		portsJSON, _ := json.Marshal(app.Ports)
		configsJSON, _ := json.Marshal(app.Configs)
		mountsJSON, _ := json.Marshal(app.Mounts)
		
		_, err = queries.CreateApp(ctx, db.CreateAppParams{
			ChartID: storedChart.ID,
			Name:    app.Name,
			Image:   sql.NullString{String: app.Image, Valid: app.Image != ""},
			AppType: sql.NullString{String: app.Type, Valid: app.Type != ""},
			Ports:   sql.NullString{String: string(portsJSON), Valid: len(app.Ports) > 0},
			Configs: sql.NullString{String: string(configsJSON), Valid: len(app.Configs) > 0},
			Mounts:  sql.NullString{String: string(mountsJSON), Valid: len(app.Mounts) > 0},
		})
		if err != nil {
			fmt.Printf("Warning: failed to store app %s: %v\n", app.Name, err)
		}
	}
	
	return &storedChart, nil
}

func getStoredCharts(c *gin.Context) {
	fmt.Printf("=== ENTERING getStoredCharts FUNCTION ===\n")
	fmt.Printf("Request URL: %s\n", c.Request.URL.String())
	fmt.Printf("Request Method: %s\n", c.Request.Method)
	fmt.Printf("User-Agent: %s\n", c.Request.Header.Get("User-Agent"))
	fmt.Printf("This should ONLY use database, NO directory scanning!\n")
	
	ctx := context.Background()
	
	// Test database connection first
	fmt.Printf("Testing database connection...\n")
	if err := database.Ping(); err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	fmt.Printf("✅ Database connection successful\n")
	
	fmt.Printf("Executing queries.ListCharts...\n")
	charts, err := queries.ListCharts(ctx)
	if err != nil {
		fmt.Printf("❌ Error fetching charts from DB: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database query failed: %v", err)})
		return
	}
	
	fmt.Printf("✅ Found %d charts in database\n", len(charts))
	
	// Initialize as empty slice to ensure we always return an array
	chartInfos := make([]ChartInfo, 0)
	
	for i, chart := range charts {
		fmt.Printf("Processing chart %d: %s (v%s)\n", i+1, chart.Name, chart.Version)
		
		// Get dependencies for this chart
		dependencies, err := queries.GetChartDependencies(ctx, chart.ID)
		if err != nil {
			fmt.Printf("Warning: failed to get dependencies for %s: %v\n", chart.Name, err)
			dependencies = []db.GetChartDependenciesRow{}
		}
		
		// Convert dependencies to Chart format
		var chartDeps []Dependency
		for _, dep := range dependencies {
			chartDeps = append(chartDeps, Dependency{
				Name:       dep.DependencyName,
				Version:    dep.DependencyVersion,
				Repository: dep.Repository.String,
				Condition:  dep.ConditionField.String,
			})
		}
		
		chartInfo := ChartInfo{
			Chart: Chart{
				Name:         chart.Name,
				Version:      chart.Version,
				Description:  chart.Description.String,
				Type:         chart.Type,
				Dependencies: chartDeps,
			},
			ImageTag:  "N/A",
			CanaryTag: "N/A",
		}
		
		if chart.ImageTag.Valid {
			chartInfo.ImageTag = chart.ImageTag.String
		}
		if chart.CanaryTag.Valid {
			chartInfo.CanaryTag = chart.CanaryTag.String
		}
		
		fmt.Printf("Chart %s has %d dependencies\n", chart.Name, len(chartDeps))
		chartInfos = append(chartInfos, chartInfo)
	}
	
	fmt.Printf("✅ Returning %d chart infos\n", len(chartInfos))
	fmt.Printf("Response data: %+v\n", chartInfos)
	fmt.Printf("=== EXITING getStoredCharts FUNCTION ===\n")
	c.JSON(http.StatusOK, chartInfos)
}

func getStoredChartInfo(c *gin.Context) {
	ctx := context.Background()
	chartName := c.Param("name")
	
	chart, err := queries.GetChart(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chart not found"})
		return
	}
	
	chartInfo := ChartInfo{
		Chart: Chart{
			Name:        chart.Name,
			Version:     chart.Version,
			Description: chart.Description.String,
			Type:        chart.Type,
		},
		ImageTag:  "N/A",
		CanaryTag: "N/A",
	}
	
	if chart.ImageTag.Valid {
		chartInfo.ImageTag = chart.ImageTag.String
	}
	if chart.CanaryTag.Valid {
		chartInfo.CanaryTag = chart.CanaryTag.String
	}
	
	c.JSON(http.StatusOK, chartInfo)
}

func getChartDependencies(c *gin.Context) {
	ctx := context.Background()
	chartName := c.Param("name")
	
	fmt.Printf("🔍 Getting dependencies for chart: %s\n", chartName)
	
	chart, err := queries.GetChart(ctx, chartName)
	if err != nil {
		fmt.Printf("❌ Chart not found: %s\n", chartName)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Chart not found",
			"chart": chartName,
		})
		return
	}
	
	dependencies, err := queries.GetChartDependencies(ctx, chart.ID)
	if err != nil {
		fmt.Printf("❌ Database error getting dependencies: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if len(dependencies) == 0 {
		fmt.Printf("ℹ️  Chart %s has no dependencies\n", chartName)
		c.JSON(http.StatusOK, gin.H{
			"message": "Chart has no dependencies",
			"chart": chartName,
			"dependencies": []interface{}{},
			"count": 0,
		})
		return
	}
	
	fmt.Printf("✅ Found %d dependencies for chart %s\n", len(dependencies), chartName)
	c.JSON(http.StatusOK, gin.H{
		"chart": chartName,
		"dependencies": dependencies,
		"count": len(dependencies),
	})
}

func deleteChart(c *gin.Context) {
	ctx := context.Background()
	chartName := c.Param("name")
	
	err := queries.DeleteChart(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Chart deleted successfully"})
}

func healthCheck(c *gin.Context) {
	fmt.Printf("GET /api/health - Health check requested\n")
	
	// Test database connection
	if err := database.Ping(); err != nil {
		fmt.Printf("Database health check failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"database": "disconnected",
			"error": err.Error(),
		})
		return
	}
	
	// Test a simple query
	ctx := context.Background()
	charts, err := queries.ListCharts(ctx)
	if err != nil {
		fmt.Printf("Database query failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"database": "connected",
			"query": "failed",
			"error": err.Error(),
		})
		return
	}
	
	fmt.Printf("Health check passed - found %d charts\n", len(charts))
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"database": "connected",
		"charts_count": len(charts),
		"version": "1.0.0",
	})
}

func fetchChartDependencies(c *gin.Context) {
	fmt.Printf("=== FETCHING CHART DEPENDENCIES ===\n")
	chartName := c.Param("name")
	ctx := context.Background()
	
	// Get the chart
	chart, err := queries.GetChart(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chart not found"})
		return
	}
	
	// Get existing dependencies
	dependencies, err := queries.GetChartDependencies(ctx, chart.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	fmt.Printf("Found %d dependencies for chart %s\n", len(dependencies), chartName)
	
	// For each dependency, try to fetch it if it doesn't exist in our database
	var fetchedCharts []ChartInfo
	
	for _, dep := range dependencies {
		// Check if dependency chart already exists
		_, err := queries.GetChart(ctx, dep.DependencyName)
		if err != nil {
			// Chart doesn't exist, try to fetch it
			fmt.Printf("Attempting to fetch dependency: %s\n", dep.DependencyName)
			
			// Try common registry patterns
			registries := []string{
				"oci://registry-1.docker.io/bitnamicharts/" + dep.DependencyName,
				"oci://registry.k8s.io/" + dep.DependencyName + "/" + dep.DependencyName,
			}
			
			for _, registryURL := range registries {
				chartInfo, err := tryFetchChart(registryURL, dep.DependencyName, dep.DependencyVersion)
				if err == nil {
					// Successfully fetched, store it
					_, storeErr := storeChartInDB(*chartInfo, []spec.App{}, registryURL)
					if storeErr == nil {
						fetchedCharts = append(fetchedCharts, *chartInfo)
						fmt.Printf("Successfully fetched and stored: %s\n", dep.DependencyName)
						break
					}
				}
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Processed %d dependencies", len(dependencies)),
		"fetched_charts": fetchedCharts,
		"total_dependencies": len(dependencies),
	})
}

func tryFetchChart(chartURL, name, version string) (*ChartInfo, error) {
	// Initialize ChartUtils
	chartUtils, err := charts.NewChartUtils()
	if err != nil {
		return nil, err
	}
	
	// Try to authenticate if Docker config exists
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "compose", "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var config DockerConfig
		if err := json.Unmarshal(data, &config); err == nil {
			authInfo := &spec.AuthInfo{
				Username: config.Username,
				Password: config.Password,
				Registry: config.Registry,
			}
			chartUtils.Authenticate(authInfo)
		}
	}
	
	// Try to parse the chart
	chartInfo, _, err := safeParseChart(chartUtils, ChartRequest{
		ChartURL: chartURL,
		ValuesPath: "values",
		SetValues: []string{},
		UseHostNetwork: false,
	})
	
	if err != nil {
		return nil, err
	}
	
	return &chartInfo, nil
}



func getDockerConfig(c *gin.Context) {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "compose", "config.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Docker config not found"})
		return
	}
	
	var config DockerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse docker config"})
		return
	}
	
	// Don't expose the password in the response for security
	response := map[string]string{
		"username": config.Username,
		"registry": config.Registry,
		"status":   "configured",
	}
	
	c.JSON(http.StatusOK, response)
}

func authenticate(c *gin.Context) {
	// Read Docker config and authenticate with Helm registry
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "compose", "config.json")
	
	fmt.Printf("Looking for config at: %s\n", configPath)
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Config read error: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Docker config not found", 
			"path": configPath,
			"details": err.Error(),
		})
		return
	}
	
	fmt.Printf("Config data: %s\n", string(data))
	
	var config DockerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Config parse error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse docker config",
			"details": err.Error(),
		})
		return
	}
	
	fmt.Printf("Parsed config: Username=%s, Registry=%s\n", config.Username, config.Registry)
	
	// Initialize ChartUtils and authenticate
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
		fmt.Printf("Authentication error: %v\n", err)
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

func fetchChart(c *gin.Context) {
	var req ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ Invalid request body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	fmt.Printf("🚀 Fetching chart: %s\n", req.ChartURL)
	
	// Initialize ChartUtils
	chartUtils, err := charts.NewChartUtils()
	if err != nil {
		fmt.Printf("❌ Failed to initialize chart utils: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize chart utils"})
		return
	}
	
	// Authenticate if Docker config exists
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "compose", "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var config DockerConfig
		if err := json.Unmarshal(data, &config); err == nil {
			authInfo := &spec.AuthInfo{
				Username: config.Username,
				Password: config.Password,
				Registry: config.Registry,
			}
			if authErr := chartUtils.Authenticate(authInfo); authErr != nil {
				fmt.Printf("⚠️  Authentication warning: %v\n", authErr)
			}
		}
	}
	
	// Try to fetch chart info safely
	chartInfo, apps, err := safeParseChart(chartUtils, req)
	if err != nil {
		fmt.Printf("❌ Failed to fetch chart %s: %v\n", req.ChartURL, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to fetch chart",
			"details": err.Error(),
			"chart_url": req.ChartURL,
		})
		return
	}
	
	fmt.Printf("✅ Successfully parsed chart: %s v%s\n", chartInfo.Chart.Name, chartInfo.Chart.Version)
	
	// Store or update chart in database
	storedChart, err := storeChartInDB(chartInfo, apps, req.ChartURL)
	if err != nil {
		fmt.Printf("⚠️  Warning: failed to store chart in database: %v\n", err)
		// Continue anyway, return the chart info
	} else {
		fmt.Printf("✅ Chart stored in database with ID: %d\n", storedChart.ID)
	}
	
	response := map[string]interface{}{
		"message": "Chart fetched successfully",
		"chart": chartInfo,
		"apps":  apps,
		"dependencies_count": len(chartInfo.Chart.Dependencies),
	}
	
	if storedChart != nil {
		response["stored"] = true
		response["chart_id"] = storedChart.ID
	}
	
	// Add info about dependencies
	if len(chartInfo.Chart.Dependencies) == 0 {
		response["info"] = "Chart has no dependencies"
	} else {
		response["info"] = fmt.Sprintf("Chart has %d dependencies", len(chartInfo.Chart.Dependencies))
	}
	
	c.JSON(http.StatusOK, response)
}

func safeParseChart(chartUtils *charts.ChartUtils, req ChartRequest) (ChartInfo, []spec.App, error) {
	// Set default values path if empty
	valuesPath := req.ValuesPath
	if valuesPath == "" {
		valuesPath = "values"
	}
	
	// First try to just template the chart to get basic info
	var rel *release.Release
	var err error
	
	// Wrap in a recovery function to catch panics
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in Template: %v\n", r)
				err = fmt.Errorf("chart templating panicked: %v", r)
			}
		}()
		rel, err = chartUtils.Template(req.ChartURL, valuesPath, req.SetValues)
	}()
	
	if err != nil {
		return ChartInfo{}, nil, fmt.Errorf("chart templating failed: %v", err)
	}
	
	if rel == nil {
		return ChartInfo{}, nil, fmt.Errorf("chart templating returned nil release")
	}
	
	// Extract basic chart info
	chartName := charts.ExtractName(req.ChartURL)
	if rel.Chart != nil && rel.Chart.Metadata != nil {
		chartName = rel.Chart.Metadata.Name
	}
	
	chartInfo := ChartInfo{
		Chart: Chart{
			Name:        chartName,
			Version:     "unknown",
			Description: fmt.Sprintf("Chart fetched from %s", req.ChartURL),
			Type:        "application",
		},
		ImageTag:  "N/A",
		CanaryTag: "N/A",
	}
	
	if rel.Chart != nil && rel.Chart.Metadata != nil {
		fmt.Printf("📋 Chart metadata found for %s\n", chartName)
		fmt.Printf("  Name: %s\n", rel.Chart.Metadata.Name)
		fmt.Printf("  Version: %s\n", rel.Chart.Metadata.Version)
		fmt.Printf("  Type: %s\n", rel.Chart.Metadata.Type)
		fmt.Printf("  Description: %s\n", rel.Chart.Metadata.Description)
		fmt.Printf("  Dependencies pointer: %p\n", rel.Chart.Metadata.Dependencies)
		
		chartInfo.Chart.Version = rel.Chart.Metadata.Version
		if rel.Chart.Metadata.Description != "" {
			chartInfo.Chart.Description = rel.Chart.Metadata.Description
		}
		if rel.Chart.Metadata.Type != "" {
			chartInfo.Chart.Type = rel.Chart.Metadata.Type
		}
		
		// Extract dependencies from chart metadata
		if rel.Chart.Metadata.Dependencies != nil {
			fmt.Printf("📦 Dependencies array exists with length: %d\n", len(rel.Chart.Metadata.Dependencies))
			if len(rel.Chart.Metadata.Dependencies) > 0 {
				fmt.Printf("✅ Found %d dependencies in chart metadata\n", len(rel.Chart.Metadata.Dependencies))
				for i, dep := range rel.Chart.Metadata.Dependencies {
					fmt.Printf("  Dependency %d: %+v\n", i+1, dep)
					chartInfo.Chart.Dependencies = append(chartInfo.Chart.Dependencies, Dependency{
						Name:       dep.Name,
						Version:    dep.Version,
						Repository: dep.Repository,
						Condition:  dep.Condition,
					})
					fmt.Printf("  📦 Added dependency: %s v%s from %s\n", dep.Name, dep.Version, dep.Repository)
				}
			} else {
				fmt.Printf("ℹ️  Dependencies array is empty for chart %s\n", chartName)
			}
		} else {
			fmt.Printf("ℹ️  No dependencies metadata found for chart %s\n", chartName)
		}
	} else {
		fmt.Printf("⚠️  No chart metadata found for %s\n", chartName)
	}
	
	// Try to extract metadata from the manifest
	if rel.Manifest != "" {
		metadata := extractManifestMetadata(rel.Manifest)
		if metadata.ImageTag != "N/A" {
			chartInfo.ImageTag = metadata.ImageTag
		}
		if metadata.CanaryTag != "N/A" {
			chartInfo.CanaryTag = metadata.CanaryTag
		}
		
		// Store manifest metadata for later database storage
		chartInfo.ManifestMetadata = &metadata
	}
	
	// Try to parse apps using the compose package, but catch panics
	var apps []spec.App
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in Parse: %v\n", r)
				apps = []spec.App{} // Return empty apps on panic
			}
		}()
		
		parsedApps, parseErr := chartUtils.Parse(req.ChartURL, valuesPath, req.SetValues, req.UseHostNetwork)
		if parseErr != nil {
			fmt.Printf("Parse error (non-fatal): %v\n", parseErr)
			apps = []spec.App{}
		} else {
			apps = parsedApps
		}
	}()
	
	return chartInfo, apps, nil
}

type ManifestMetadata struct {
	ImageTag       string   `json:"imageTag"`
	CanaryTag      string   `json:"canaryTag"`
	ContainerImages []string `json:"containerImages"`
	IngressPaths   []string `json:"ingressPaths"`
	ServicePorts   []string `json:"servicePorts"`
}

func extractManifestMetadata(manifest string) ManifestMetadata {
	metadata := ManifestMetadata{
		ImageTag:        "N/A",
		CanaryTag:       "N/A",
		ContainerImages: []string{},
		IngressPaths:    []string{},
		ServicePorts:    []string{},
	}
	
	// Split manifest into resources
	resources := strings.Split(manifest, "---")
	
	for _, resource := range resources {
		resource = strings.TrimSpace(resource)
		if resource == "" {
			continue
		}
		
		lines := strings.Split(resource, "\n")
		var currentKind string
		var inSpec, inIngress, inService bool
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			
			// Detect resource kind
			if strings.HasPrefix(line, "kind:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentKind = strings.TrimSpace(parts[1])
					inIngress = currentKind == "Ingress"
					inService = currentKind == "Service"
				}
			}
			
			// Track if we're in spec section
			if strings.HasPrefix(line, "spec:") {
				inSpec = true
			}
			
			// Extract container images
			if strings.Contains(line, "image:") && !strings.Contains(line, "imagePullPolicy") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					imageRef := strings.TrimSpace(strings.Join(parts[1:], ":"))
					imageRef = strings.Trim(imageRef, "\"'")
					
					// Add to container images if not already present
					found := false
					for _, existing := range metadata.ContainerImages {
						if existing == imageRef {
							found = true
							break
						}
					}
					if !found && imageRef != "" {
						metadata.ContainerImages = append(metadata.ContainerImages, imageRef)
					}
					
					// Extract tag for primary image tag
					if strings.Contains(imageRef, ":") {
						tagParts := strings.Split(imageRef, ":")
						tag := tagParts[len(tagParts)-1]
						if metadata.ImageTag == "N/A" {
							metadata.ImageTag = tag
						}
						
						// Check for canary tags
						if strings.Contains(strings.ToLower(tag), "canary") && metadata.CanaryTag == "N/A" {
							metadata.CanaryTag = tag
						}
					}
				}
			}
			
			// Extract ingress paths
			if inIngress && strings.Contains(line, "path:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					path := strings.TrimSpace(parts[1])
					path = strings.Trim(path, "\"'")
					if path != "" && path != "/" {
						metadata.IngressPaths = append(metadata.IngressPaths, path)
					}
				}
			}
			
			// Extract service ports
			if inService && inSpec && strings.Contains(line, "port:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					port := strings.TrimSpace(parts[1])
					port = strings.Trim(port, "\"'")
					if port != "" {
						metadata.ServicePorts = append(metadata.ServicePorts, port)
					}
				}
			}
		}
	}
	
	return metadata
}

func extractImageTagsFromManifest(manifest string) (string, string) {
	metadata := extractManifestMetadata(manifest)
	return metadata.ImageTag, metadata.CanaryTag
}

func testChart(c *gin.Context) {
	var req ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Initialize ChartUtils
	chartUtils, err := charts.NewChartUtils()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize chart utils"})
		return
	}
	
	// Authenticate if Docker config exists
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "compose", "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var config DockerConfig
		if err := json.Unmarshal(data, &config); err == nil {
			authInfo := &spec.AuthInfo{
				Username: config.Username,
				Password: config.Password,
				Registry: config.Registry,
			}
			if authErr := chartUtils.Authenticate(authInfo); authErr != nil {
				fmt.Printf("Authentication warning: %v\n", authErr)
			}
		}
	}
	
	// Just try to template the chart without parsing
	valuesPath := req.ValuesPath
	if valuesPath == "" {
		valuesPath = "values"
	}
	
	rel, err := chartUtils.Template(req.ChartURL, valuesPath, req.SetValues)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to template chart: %v", err)})
		return
	}
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"name":     rel.Name,
		"manifest": rel.Manifest[:min(1000, len(rel.Manifest))], // First 1000 chars
		"chart":    rel.Chart.Metadata.Name,
		"version":  rel.Chart.Metadata.Version,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getChartVersions(c *gin.Context) {
	ctx := context.Background()
	chartName := c.Param("name")
	
	versions, err := queries.ListChartVersions(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"chart": chartName,
		"versions": versions,
		"count": len(versions),
	})
}

func switchChartVersion(c *gin.Context) {
	ctx := context.Background()
	chartName := c.Param("name")
	
	var req struct {
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Mark all versions as not latest
	err := queries.SetLatestVersion(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Set the requested version as latest
	err = queries.SetVersionAsLatest(ctx, db.SetVersionAsLatestParams{
		Name:    chartName,
		Version: req.Version,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Switched %s to version %s", chartName, req.Version),
		"chart": chartName,
		"version": req.Version,
	})
}

func deleteChartVersion(c *gin.Context) {
	ctx := context.Background()
	chartName := c.Param("name")
	version := c.Param("version")
	
	err := queries.DeleteChartVersion(ctx, db.DeleteChartVersionParams{
		Name:    chartName,
		Version: version,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Deleted version %s of chart %s", version, chartName),
	})
}

func fetchChartDependenciesAdvanced(c *gin.Context) {
	fmt.Printf("=== ADVANCED DEPENDENCY FETCHING ===\n")
	chartName := c.Param("name")
	ctx := context.Background()
	
	// Get the chart
	_, err := queries.GetChart(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chart not found"})
		return
	}
	
	// Get dependencies using the new query
	dependencies, err := queries.FetchChartDependencies(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	fmt.Printf("Found %d dependencies for chart %s\n", len(dependencies), chartName)
	
	var fetchedCharts []ChartInfo
	var errors []string
	
	for _, dep := range dependencies {
		fmt.Printf("Processing dependency: %s v%s\n", dep.DependencyName, dep.DependencyVersion)
		
		// Check if dependency chart already exists
		_, err := queries.GetChart(ctx, dep.DependencyName)
		if err != nil {
			// Chart doesn't exist, try to fetch it
			fmt.Printf("Attempting to fetch dependency: %s\n", dep.DependencyName)
			
			// Try different registry patterns
			registries := []string{
				"oci://registry-1.docker.io/bitnamicharts/" + dep.DependencyName,
				"oci://registry.k8s.io/" + dep.DependencyName + "/" + dep.DependencyName,
				"oci://ghcr.io/helm/" + dep.DependencyName,
			}
			
			// If repository is specified, try that first
			if dep.Repository.Valid && dep.Repository.String != "" {
				registries = append([]string{dep.Repository.String}, registries...)
			}
			
			var chartInfo *ChartInfo
			var fetchErr error
			
			for _, registryURL := range registries {
				fmt.Printf("Trying registry: %s\n", registryURL)
				chartInfo, fetchErr = tryFetchChart(registryURL, dep.DependencyName, dep.DependencyVersion)
				if fetchErr == nil {
					break
				}
				fmt.Printf("Failed to fetch from %s: %v\n", registryURL, fetchErr)
			}
			
			if chartInfo != nil {
				// Successfully fetched, store it
				_, storeErr := storeChartInDB(*chartInfo, []spec.App{}, registries[0])
				if storeErr == nil {
					fetchedCharts = append(fetchedCharts, *chartInfo)
					fmt.Printf("✅ Successfully fetched and stored: %s v%s\n", dep.DependencyName, chartInfo.Chart.Version)
				} else {
					errors = append(errors, fmt.Sprintf("Failed to store %s: %v", dep.DependencyName, storeErr))
				}
			} else {
				errors = append(errors, fmt.Sprintf("Failed to fetch %s from any registry: %v", dep.DependencyName, fetchErr))
			}
		} else {
			fmt.Printf("Dependency %s already exists in database\n", dep.DependencyName)
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Processed %d dependencies", len(dependencies)),
		"fetched_charts": fetchedCharts,
		"total_dependencies": len(dependencies),
		"newly_fetched": len(fetchedCharts),
		"errors": errors,
	})
}

// Update chart manifest metadata
func updateChartManifestMetadata(chartID int64, metadata ManifestMetadata) error {
	ctx := context.Background()
	
	// Convert arrays to JSON
	containerImagesJSON, _ := json.Marshal(metadata.ContainerImages)
	ingressPathsJSON, _ := json.Marshal(metadata.IngressPaths)
	servicePortsJSON, _ := json.Marshal(metadata.ServicePorts)
	
	// Update the chart with manifest metadata
	_, err := database.ExecContext(ctx, `
		UPDATE charts 
		SET ingress_paths = ?, 
		    container_images = ?, 
		    service_ports = ?,
		    manifest_parsed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, string(ingressPathsJSON), string(containerImagesJSON), string(servicePortsJSON), chartID)
	
	if err != nil {
		return fmt.Errorf("failed to update manifest metadata: %v", err)
	}
	
	fmt.Printf("✅ Updated manifest metadata for chart ID %d\n", chartID)
	return nil
}

// Get chart manifest metadata
func getChartManifestMetadata(chartID int64) (*ManifestMetadata, error) {
	ctx := context.Background()
	
	var ingressPathsJSON, containerImagesJSON, servicePortsJSON sql.NullString
	err := database.QueryRowContext(ctx, `
		SELECT ingress_paths, container_images, service_ports 
		FROM charts 
		WHERE id = ?
	`, chartID).Scan(&ingressPathsJSON, &containerImagesJSON, &servicePortsJSON)
	
	if err != nil {
		return nil, err
	}
	
	metadata := &ManifestMetadata{
		ContainerImages: []string{},
		IngressPaths:    []string{},
		ServicePorts:    []string{},
	}
	
	// Parse JSON arrays
	if ingressPathsJSON.Valid && ingressPathsJSON.String != "" {
		json.Unmarshal([]byte(ingressPathsJSON.String), &metadata.IngressPaths)
	}
	
	if containerImagesJSON.Valid && containerImagesJSON.String != "" {
		json.Unmarshal([]byte(containerImagesJSON.String), &metadata.ContainerImages)
	}
	
	if servicePortsJSON.Valid && servicePortsJSON.String != "" {
		json.Unmarshal([]byte(servicePortsJSON.String), &metadata.ServicePorts)
	}
	
	return metadata, nil
}

// Registry configuration struct
type RegistryConfig struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	RegistryURL string `json:"registry_url" db:"registry_url"`
	Username    string `json:"username" db:"username"`
	Password    string `json:"password" db:"password"`
	IsDefault   bool   `json:"is_default" db:"is_default"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

// Get all registry configurations
func getRegistryConfigs(c *gin.Context) {
	ctx := context.Background()
	
	rows, err := database.QueryContext(ctx, `
		SELECT id, name, registry_url, username, password, is_default, created_at, updated_at 
		FROM registry_configs 
		ORDER BY is_default DESC, name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	
	var configs []RegistryConfig
	for rows.Next() {
		var config RegistryConfig
		err := rows.Scan(&config.ID, &config.Name, &config.RegistryURL, &config.Username, 
			&config.Password, &config.IsDefault, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		configs = append(configs, config)
	}
	
	c.JSON(http.StatusOK, configs)
}

// Create new registry configuration
func createRegistryConfig(c *gin.Context) {
	ctx := context.Background()
	
	var config RegistryConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// If this is set as default, unset other defaults
	if config.IsDefault {
		_, err := database.ExecContext(ctx, "UPDATE registry_configs SET is_default = FALSE")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing defaults"})
			return
		}
	}
	
	result, err := database.ExecContext(ctx, `
		INSERT INTO registry_configs (name, registry_url, username, password, is_default) 
		VALUES (?, ?, ?, ?, ?)
	`, config.Name, config.RegistryURL, config.Username, config.Password, config.IsDefault)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	id, _ := result.LastInsertId()
	config.ID = id
	
	c.JSON(http.StatusCreated, config)
}

// Update registry configuration
func updateRegistryConfig(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	
	var config RegistryConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// If this is set as default, unset other defaults
	if config.IsDefault {
		_, err := database.ExecContext(ctx, "UPDATE registry_configs SET is_default = FALSE WHERE id != ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing defaults"})
			return
		}
	}
	
	_, err := database.ExecContext(ctx, `
		UPDATE registry_configs 
		SET name = ?, registry_url = ?, username = ?, password = ?, is_default = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, config.Name, config.RegistryURL, config.Username, config.Password, config.IsDefault, id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, config)
}

// Delete registry configuration
func deleteRegistryConfig(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	
	_, err := database.ExecContext(ctx, "DELETE FROM registry_configs WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Registry configuration deleted"})
}

// Set registry as default
func setDefaultRegistry(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	
	// Unset all defaults first
	_, err := database.ExecContext(ctx, "UPDATE registry_configs SET is_default = FALSE")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update existing defaults"})
		return
	}
	
	// Set this one as default
	_, err = database.ExecContext(ctx, "UPDATE registry_configs SET is_default = TRUE WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Default registry updated"})
}
package server

import (
	"chartpaper/internal/db"
	"chartpaper/pkg"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ashupednekar/compose/pkg/charts"
	"github.com/ashupednekar/compose/pkg/spec"
	"github.com/gin-gonic/gin"
)

func (s *Server) fetchChart(c *gin.Context) {
	var req pkg.ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ Invalid request body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	fmt.Printf("🚀 Fetching chart: %s\n", req.ChartURL)
	
	chartUtils, err := charts.NewChartUtils()
	if err != nil {
		fmt.Printf("❌ Failed to initialize chart utils: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize chart utils"})
		return
	}
	
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "compose", "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var config pkg.DockerConfig
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
	
	chartInfo, apps, err := pkg.SafeParseChart(chartUtils, req)
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
	
	storedChart, err := storeChartInDB(s.db, chartInfo, apps, req.ChartURL)
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



func (s *Server) getStoredCharts(c *gin.Context) {
	queries := db.New(s.db)
	log.Printf("=== ENTERING getStoredCharts FUNCTION ===\n")
	log.Printf("Request URL: %s\n", c.Request.URL.String())
	log.Printf("Request Method: %s\n", c.Request.Method)
	log.Printf("User-Agent: %s\n", c.Request.Header.Get("User-Agent"))
	log.Printf("This should ONLY use database, NO directory scanning!\n")
	
	ctx := context.Background()
	
	log.Printf("Testing database connection...\n")
	if err := s.db.Ping(); err != nil {
		log.Printf("❌ Database connection failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	log.Printf("✅ Database connection successful\n")
	
	log.Printf("Executing queries.ListCharts...\n")
	charts, err := queries.ListCharts(ctx)
	if err != nil {
		log.Printf("❌ Error fetching charts from DB: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database query failed: %v", err)})
		return
	}
	
	log.Printf("✅ Found %d charts in database\n", len(charts))
	
	// Initialize as empty slice to ensure we always return an array
	chartInfos := make([]pkg.ChartInfo, 0)
	
	for i, chart := range charts {
		log.Printf("Processing chart %d: %s (v%s)\n", i+1, chart.Name, chart.Version)
		
		// Get dependencies for this chart
		dependencies, err := queries.GetChartDependencies(ctx, int32(chart.ID))
		if err != nil {
			log.Printf("Warning: failed to get dependencies for %s: %v\n", chart.Name, err)
			dependencies = []db.GetChartDependenciesRow{}
		}
		
		// Convert dependencies to Chart format
		var chartDeps []pkg.Dependency
		for _, dep := range dependencies {
			chartDeps = append(chartDeps, pkg.Dependency{
				Name:       dep.DependencyName,
				Version:    dep.DependencyVersion,
				Repository: dep.Repository.String,
				Condition:  dep.ConditionField.String,
			})
		}
		
		chartInfo := pkg.ChartInfo{
			Chart: pkg.Chart{
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
		
		log.Printf("Chart %s has %d dependencies\n", chart.Name, len(chartDeps))
		chartInfos = append(chartInfos, chartInfo)
	}
	
	log.Printf("✅ Returning %d chart infos\n", len(chartInfos))
	log.Printf("Response data: %+v\n", chartInfos)
	log.Printf("=== EXITING getStoredCharts FUNCTION ===\n")
	c.JSON(http.StatusOK, chartInfos)
}

func (s *Server) getStoredChartInfo(c *gin.Context) {
	queries := db.New(s.db)
	ctx := context.Background()
	chartName := c.Param("name")
	
	chart, err := queries.GetChart(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chart not found"})
		return
	}
	
	chartInfo := pkg.ChartInfo{
		Chart: pkg.Chart{
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

func (s *Server) deleteChart(c *gin.Context) {
	queries := db.New(s.db)
	ctx := context.Background()
	chartName := c.Param("name")
	
	err := queries.DeleteChart(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Chart deleted successfully"})
}

func (s *Server) deleteChartVersion(c *gin.Context) {
	queries := db.New(s.db)
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

	c.JSON(http.StatusOK, gin.H{"message": "Chart version deleted successfully"})
}

func (s *Server) getChartVersions(c *gin.Context) {
	ctx := context.Background()
	chartName := c.Param("name")
	queries := db.New(s.db)
	
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

func (s *Server) switchChartVersion(c *gin.Context) {
	ctx := context.Background()
	chartName := c.Param("name")
	queries := db.New(s.db)
	
	var req struct {
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
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


func storeChartInDB(database *sql.DB, chartInfo pkg.ChartInfo, apps []spec.App, chartURL string) (*db.Chart, error) {
	ctx := context.Background()
	queries := db.New(database)
	existingChart, err := queries.GetChart(ctx, chartInfo.Chart.Name)
	
	var storedChart db.Chart
	
	if err != nil {
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
			err = pkg.UpdateChartManifestMetadata(database, int64(storedChart.ID), *chartInfo.ManifestMetadata)
			if err != nil {
				fmt.Printf("⚠️  Warning: failed to update manifest metadata: %v\n", err)
			}
		}
	} else {
		// Chart exists, use the existing one
		fmt.Printf("📝 Using existing chart: %s (ID: %d)\n", chartInfo.Chart.Name, existingChart.ID)
		storedChart = existingChart
		
		// Clear existing dependencies
		queries.DeleteChartDependencies(ctx, int32(storedChart.ID))
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
			if depChartInfo, depErr := pkg.TryFetchChart(depChartURL, dep.Name, dep.Version); depErr == nil {
				depImageTag = depChartInfo.ImageTag
				depCanaryTag = depChartInfo.CanaryTag
				fmt.Printf("✅ Got dependency tags: image=%s, canary=%s\n", depImageTag, depCanaryTag)
			} else {
				fmt.Printf("⚠️ Could not fetch dependency info: %v\n", depErr)
			}
		}
		
		// For now, store dependency without tags until we can update the schema properly
		depResult, err := queries.CreateDependency(ctx, db.CreateDependencyParams{
			ChartID:           int32(storedChart.ID),
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
			ChartID: int32(storedChart.ID),
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

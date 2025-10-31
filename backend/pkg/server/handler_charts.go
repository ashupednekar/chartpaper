package server

import (
	"chartpaper/internal/db"
	"chartpaper/pkg"
	"context"
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

func (s *Server) fetchChart(c *gin.Context) {
	var req pkg.ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Invalid request body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	log.Printf("🚀 Fetching chart: %s\n", req.ChartURL)
	
	chartUtils, err := charts.NewChartUtils(true)
	if err != nil {
		log.Printf("❌ Failed to initialize chart utils: %v\n", err)
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
				log.Printf("⚠️  Authentication warning: %v\n", authErr)
			}
		}
	}
	
	chartInfo, apps, err := pkg.SafeParseChart(chartUtils, req)
	if err != nil {
		log.Printf("❌ Failed to fetch chart %s: %v\n", req.ChartURL, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to fetch chart",
			"details": err.Error(),
			"chart_url": req.ChartURL,
		})
		return
	}
	
	log.Printf("✅ Successfully parsed chart: %s v%s\n", chartInfo.Chart.Name, chartInfo.Chart.Version)
	
	storedChart, err := pkg.StoreChartInDB(s.db, chartInfo, apps, req.ChartURL)
	if err != nil {
		log.Printf("⚠️  Warning: failed to store chart in database: %v\n", err)
		// Continue anyway, return the chart info
	} else {
		log.Printf("✅ Chart stored in database with ID: %d\n", storedChart.ID)
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
	if err := s.db.Ping(ctx); err != nil {
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
			var repo, cond string
			if dep.Repository.Valid {
				repo = dep.Repository.String
			}
			if dep.ConditionField.Valid {
				cond = dep.ConditionField.String
			}
			chartDeps = append(chartDeps, pkg.Dependency{
				Name:       dep.DependencyName,
				Version:    dep.DependencyVersion,
				Repository: repo,
				Condition:  cond,
			})
		}
		
		var desc string
		if chart.Description.Valid {
			desc = chart.Description.String
		}
		chartInfo := pkg.ChartInfo{
			Chart: pkg.Chart{
				Name:         chart.Name,
				Version:      chart.Version,
				Description:  desc,
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
	
	var desc string
	if chart.Description.Valid {
		desc = chart.Description.String
	}
	chartInfo := pkg.ChartInfo{
		Chart: pkg.Chart{
			Name:        chart.Name,
			Version:     chart.Version,
			Description: desc,
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


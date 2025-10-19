package server

import (
	"chartpaper/internal/db"
	"chartpaper/pkg"
	"context"
	"fmt"
	"net/http"

	"github.com/ashupednekar/compose/pkg/spec"
	"github.com/gin-gonic/gin"
)

func (s *Server) fetchChartDependencies(c *gin.Context) {
	queries := db.New(s.db)
	fmt.Printf("=== FETCHING CHART DEPENDENCIES ===\n")
	chartName := c.Param("name")
	ctx := context.Background()


	chart, err := queries.GetChart(ctx, chartName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chart not found"})
		return
	}


	dependencies, err := queries.GetChartDependencies(ctx, int32(chart.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Found %d dependencies for chart %s\n", len(dependencies), chartName)


	var fetchedCharts []pkg.ChartInfo

	for _, dep := range dependencies {

		_, err := queries.GetChart(ctx, dep.DependencyName)
		if err != nil {

			fmt.Printf("Attempting to fetch dependency: %s\n", dep.DependencyName)


			registries := []string{
				"oci://registry-1.docker.io/bitnamicharts/" + dep.DependencyName,
				"oci://registry.k8s.io/" + dep.DependencyName + "/" + dep.DependencyName,
			}

			for _, registryURL := range registries {
				chartInfo, err := pkg.TryFetchChart(registryURL, dep.DependencyName, dep.DependencyVersion)
				if err == nil {

					_, storeErr := pkg.StoreChartInDB(s.db, *chartInfo, []spec.App{}, registryURL)
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
		"message":            fmt.Sprintf("Processed %d dependencies", len(dependencies)),
		"fetched_charts":     fetchedCharts,
		"total_dependencies": len(dependencies),
	})
}


func (s *Server) getChartDependencies(c *gin.Context) {
	queries := db.New(s.db)
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
	
	dependencies, err := queries.GetChartDependencies(ctx, int32(chart.ID))
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

package pkg

import (
	"chartpaper/internal/db"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ashupednekar/compose/pkg/charts"
	"github.com/ashupednekar/compose/pkg/spec"
	"helm.sh/helm/v3/pkg/release"
)

func TryFetchChart(chartURL, name, version string) (*ChartInfo, error) {
	
	chartUtils, err := charts.NewChartUtils()
	if err != nil {
		return nil, err
	}
	
	
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
	
	
	chartInfo, _, err := SafeParseChart(chartUtils, ChartRequest{
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

func SafeParseChart(chartUtils *charts.ChartUtils, req ChartRequest) (ChartInfo, []spec.App, error) {
	
	valuesPath := req.ValuesPath
	if valuesPath == "" {
		valuesPath = "values"
	}
	
	
	var rel *release.Release
	var err error
	
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in Template: %v\n", r)
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
		log.Printf("📋 Chart metadata found for %s\n", chartName)
		log.Printf("  Name: %s\n", rel.Chart.Metadata.Name)
		log.Printf("  Version: %s\n", rel.Chart.Metadata.Version)
		log.Printf("  Type: %s\n", rel.Chart.Metadata.Type)
		log.Printf("  Description: %s\n", rel.Chart.Metadata.Description)
		log.Printf("  Dependencies pointer: %p\n", rel.Chart.Metadata.Dependencies)
		
		chartInfo.Chart.Version = rel.Chart.Metadata.Version
		if rel.Chart.Metadata.Description != "" {
			chartInfo.Chart.Description = rel.Chart.Metadata.Description
		}
		if rel.Chart.Metadata.Type != "" {
			chartInfo.Chart.Type = rel.Chart.Metadata.Type
		}
		
		
		if rel.Chart.Metadata.Dependencies != nil {
			log.Printf("📦 Dependencies array exists with length: %d\n", len(rel.Chart.Metadata.Dependencies))
			if len(rel.Chart.Metadata.Dependencies) > 0 {
				log.Printf("✅ Found %d dependencies in chart metadata\n", len(rel.Chart.Metadata.Dependencies))
				for i, dep := range rel.Chart.Metadata.Dependencies {
					log.Printf("  Dependency %d: %+v\n", i+1, dep)
					chartInfo.Chart.Dependencies = append(chartInfo.Chart.Dependencies, Dependency{
						Name:       dep.Name,
						Version:    dep.Version,
						Repository: dep.Repository,
						Condition:  dep.Condition,
					})
					log.Printf("  📦 Added dependency: %s v%s from %s\n", dep.Name, dep.Version, dep.Repository)
				}
			} else {
				log.Printf("ℹ️  Dependencies array is empty for chart %s\n", chartName)
			}
		} else {
			log.Printf("ℹ️  No dependencies metadata found for chart %s\n", chartName)
		}
	} else {
		log.Printf("⚠️  No chart metadata found for %s\n", chartName)
	}
	
	
	if rel.Manifest != "" {
		metadata := extractManifestMetadata(rel.Manifest)
		if metadata.ImageTag != "N/A" {
			chartInfo.ImageTag = metadata.ImageTag
		}
		if metadata.CanaryTag != "N/A" {
			chartInfo.CanaryTag = metadata.CanaryTag
		}
		
		
		chartInfo.ManifestMetadata = &metadata
	}
	
	
	var apps []spec.App
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in Parse: %v\n", r)
				apps = []spec.App{} 
			}
		}()
		
		parsedApps, parseErr := chartUtils.Parse(req.ChartURL, valuesPath, req.SetValues, req.UseHostNetwork)
		if parseErr != nil {
			log.Printf("Parse error (non-fatal): %v\n", parseErr)
			apps = []spec.App{}
		} else {
			apps = parsedApps
		}
	}()
	
	return chartInfo, apps, nil
}

func extractManifestMetadata(manifest string) ManifestMetadata {
	metadata := ManifestMetadata{
		ImageTag:        "N/A",
		CanaryTag:       "N/A",
		ContainerImages: []string{},
		IngressPaths:    []string{},
		ServicePorts:    []string{},
	}
	
	
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
			
			
			if strings.HasPrefix(line, "kind:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentKind = strings.TrimSpace(parts[1])
					inIngress = currentKind == "Ingress"
					inService = currentKind == "Service"
				}
			}
			
			
			if strings.HasPrefix(line, "spec:") {
				inSpec = true
			}
			
			
			if strings.Contains(line, "image:") && !strings.Contains(line, "imagePullPolicy") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					imageRef := strings.TrimSpace(strings.Join(parts[1:], ":"))
					imageRef = strings.Trim(imageRef, "\"'")
					
					
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
					
					
					if strings.Contains(imageRef, ":") {
						tagParts := strings.Split(imageRef, ":")
						tag := tagParts[len(tagParts)-1]
						if metadata.ImageTag == "N/A" {
							metadata.ImageTag = tag
						}
						
						
						if strings.Contains(strings.ToLower(tag), "canary") && metadata.CanaryTag == "N/A" {
							metadata.CanaryTag = tag
						}
					}
				}
			}
			
			
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

func StoreChartInDB(database *sql.DB, chartInfo ChartInfo, apps []spec.App, chartURL string) (*db.Chart, error) {
	ctx := context.Background()
	queries := db.New(database)
	
	// Simple approach: always try to get the existing chart first
	existingChart, err := queries.GetChart(ctx, chartInfo.Chart.Name)
	
	var storedChart db.Chart
	
	if err != nil {
		// Chart doesn't exist, create new one
		log.Printf("📝 Creating new chart: %s\n", chartInfo.Chart.Name)
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
			err = UpdateChartManifestMetadata(database, int64(storedChart.ID), *chartInfo.ManifestMetadata)
			if err != nil {
				log.Printf("⚠️  Warning: failed to update manifest metadata: %v\n", err)
			}
		}
	} else {
		// Chart exists, use the existing one
		log.Printf("📝 Using existing chart: %s (ID: %d)\n", chartInfo.Chart.Name, existingChart.ID)
		storedChart = existingChart
		
		// Clear existing dependencies
		queries.DeleteChartDependencies(ctx, int32(storedChart.ID))
	}
	
	log.Printf("🚀 REACHED DEPENDENCY STORAGE SECTION!\n")
	log.Printf("🔍 DEBUG: storedChart pointer: %p\n", &storedChart)
	log.Printf("🔍 DEBUG: storedChart ID: %d\n", storedChart.ID)
	
	// Store dependencies
	log.Printf("🔍 DEBUG: About to store %d dependencies for chart %s (ID: %d)\n", len(chartInfo.Chart.Dependencies), chartInfo.Chart.Name, storedChart.ID)
	log.Printf("🔍 DEBUG: Dependencies array: %+v\n", chartInfo.Chart.Dependencies)
	for i, dep := range chartInfo.Chart.Dependencies {
		log.Printf("🔍 DEBUG: Processing dependency %d: %+v\n", i+1, dep)
		
		// Try to fetch dependency chart info to get image/canary tags
		var depImageTag, depCanaryTag string = "N/A", "N/A"
		
		if dep.Repository != "" {
			// Try to fetch the dependency chart to get its image tags
			depChartURL := dep.Repository
			if !strings.HasSuffix(depChartURL, "/"+dep.Name) {
				depChartURL = depChartURL + "/" + dep.Name
			}
			
			log.Printf("🔍 Attempting to fetch dependency info from: %s\n", depChartURL)
			if depChartInfo, depErr := TryFetchChart(depChartURL, dep.Name, dep.Version); depErr == nil {
				depImageTag = depChartInfo.ImageTag
				depCanaryTag = depChartInfo.CanaryTag
				log.Printf("✅ Got dependency tags: image=%s, canary=%s\n", depImageTag, depCanaryTag)
			} else {
				log.Printf("⚠️ Could not fetch dependency info: %v\n", depErr)
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
			log.Printf("❌ ERROR: failed to store dependency %s: %v\n", dep.Name, err)
		} else {
			log.Printf("✅ SUCCESS: Stored dependency: %s v%s (ID: %d)\n", dep.Name, dep.Version, depResult.ID)
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
			log.Printf("Warning: failed to store app %s: %v\n", app.Name, err)
		}
	}
	
	return &storedChart, nil
}

func UpdateChartManifestMetadata(database *sql.DB, chartID int64, metadata ManifestMetadata) error {
	ctx := context.Background()
	
	// Convert arrays to JSON
	containerImagesJSON, _ := json.Marshal(metadata.ContainerImages)
	ingressPathsJSON, _ := json.Marshal(metadata.IngressPaths)
	servicePortsJSON, _ := json.Marshal(metadata.ServicePorts)
	
	// Update the chart with manifest metadata
	_, err := database.ExecContext(ctx, `
		UPDATE charts 
		SET ingress_paths = $1, 
		    container_images = $2, 
		    service_ports = $3,
		    manifest_parsed_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`, string(ingressPathsJSON), string(containerImagesJSON), string(servicePortsJSON), int32(chartID))
	
	if err != nil {
		return fmt.Errorf("failed to update manifest metadata: %v", err)
	}
	
	fmt.Printf("✅ Updated manifest metadata for chart ID %d\n", chartID)
	return nil
}

# Chart Paper

A tool to visualize Helm chart dependencies, library charts, and image versions in an interactive canvas.

## Features

- 📊 Visual dependency graph of Helm charts
- 📦 Library chart relationships
- 🏷️ Version tracking from Chart.yaml
- 🐳 Image tag extraction from values.yaml (.image.tag, .canary.tag)
- 🎨 Interactive canvas visualization

## Architecture

- **Backend**: Go with Gin framework
- **Frontend**: React with shadcn/ui components
- **Visualization**: Canvas-based dependency graph

## Getting Started

### Backend
```bash
cd backend
go mod tidy
go run main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

## API Endpoints

- `GET /api/charts` - List all charts
- `GET /api/charts/:name/dependencies` - Get chart dependencies
- `GET /api/charts/:name/versions` - Get version info and image tags
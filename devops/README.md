# Chart Paper DevOps

This directory contains Docker configurations and deployment scripts for Chart Paper.

## 🚀 Quick Start

### Prerequisites
- Docker (20.10+)
- Docker Compose (2.0+)

### Deploy Chart Paper
```bash
# Make the deployment script executable
chmod +x devops/deploy.sh

# Deploy Chart Paper
./devops/deploy.sh deploy
```

Chart Paper will be available at:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8000

## 📁 Files Overview

### Docker Files
- `Dockerfile.backend` - Multi-stage Docker build for Go backend
- `Dockerfile.frontend` - Multi-stage Docker build for React frontend
- `nginx.conf` - Nginx configuration with API proxy and security headers
- `docker-compose.yaml` - Complete orchestration setup

### Scripts
- `deploy.sh` - Comprehensive deployment and management script

## 🛠 Management Commands

```bash
# Deploy (build and start)
./devops/deploy.sh deploy

# View logs
./devops/deploy.sh logs

# Check status
./devops/deploy.sh status

# Restart services
./devops/deploy.sh restart

# Stop services
./devops/deploy.sh stop

# Complete cleanup (removes volumes)
./devops/deploy.sh cleanup

# Show help
./devops/deploy.sh help
```

## 🏗 Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │     Nginx       │    │    Backend      │
│   (React)       │───▶│   (Reverse      │───▶│     (Go)        │
│   Port: 80      │    │    Proxy)       │    │   Port: 8000    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                               ┌─────────────────┐
                                               │   SQLite DB     │
                                               │   (Persistent   │
                                               │    Volume)      │
                                               └─────────────────┘
```

## 🔧 Configuration

### Environment Variables

#### Backend
- `GIN_MODE=release` - Production mode
- `DB_PATH=/data/charts.db` - Database file path
- `PORT=8000` - Server port

#### Frontend
- Served by Nginx on port 80
- API requests proxied to backend

### Volumes
- `chartpaper-data` - Persistent storage for SQLite database
- `/var/run/docker.sock` - Docker socket access for registry operations

### Networks
- `chartpaper-network` - Internal bridge network for service communication

## 🔒 Security Features

### Nginx Security Headers
- X-Frame-Options: SAMEORIGIN
- X-XSS-Protection: 1; mode=block
- X-Content-Type-Options: nosniff
- Content-Security-Policy: Restrictive policy
- Referrer-Policy: no-referrer-when-downgrade

### CORS Configuration
- Proper CORS headers for API requests
- Preflight request handling
- Secure cross-origin communication

## 📊 Monitoring

### Health Checks
Both services include health checks:
- **Backend**: `GET /api/health`
- **Frontend**: HTTP 200 on port 80

### Logs
```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend
docker-compose logs -f frontend
```

## 🚀 Production Deployment

### Recommended Production Setup

1. **Use a reverse proxy** (Traefik, Nginx, or cloud load balancer)
2. **Enable HTTPS** with SSL certificates
3. **Set up monitoring** (Prometheus, Grafana)
4. **Configure log aggregation** (ELK stack, Fluentd)
5. **Set up backups** for the SQLite database

### Production Environment Variables
```bash
# Backend
GIN_MODE=release
DB_PATH=/data/charts.db
PORT=8000

# Add your registry credentials if needed
DOCKER_CONFIG_PATH=/path/to/docker/config
```

### Production Docker Compose Override
Create `docker-compose.prod.yaml`:
```yaml
version: '3.8'
services:
  frontend:
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./ssl:/etc/nginx/ssl:ro
  
  backend:
    environment:
      - GIN_MODE=release
    volumes:
      - /path/to/production/data:/data
```

Deploy with:
```bash
docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml up -d
```

## 🔧 Troubleshooting

### Common Issues

#### Port Conflicts
If ports 3000 or 8000 are in use:
```bash
# Check what's using the ports
lsof -i :3000
lsof -i :8000

# Modify docker-compose.yaml ports section
```

#### Permission Issues
```bash
# Fix Docker socket permissions
sudo chmod 666 /var/run/docker.sock

# Or add user to docker group
sudo usermod -aG docker $USER
```

#### Database Issues
```bash
# Check database volume
docker volume inspect chartpaper_chartpaper-data

# Backup database
docker cp chartpaper-backend:/data/charts.db ./backup.db
```

### Debug Mode
For development debugging:
```bash
# Run with debug output
GIN_MODE=debug docker-compose up

# Access container shell
docker exec -it chartpaper-backend sh
docker exec -it chartpaper-frontend sh
```

## 📈 Scaling

For high-traffic deployments:

1. **Use multiple backend replicas**
2. **Add Redis for session storage**
3. **Use PostgreSQL instead of SQLite**
4. **Implement horizontal pod autoscaling**
5. **Add CDN for static assets**

## 🤝 Contributing

When modifying Docker configurations:
1. Test locally with `./deploy.sh deploy`
2. Verify health checks pass
3. Test API functionality
4. Update documentation if needed

## 📝 License

Same as the main Chart Paper project.
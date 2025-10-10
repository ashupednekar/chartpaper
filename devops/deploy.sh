#!/bin/bash

# Chart Paper Deployment Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
}

# Function to build and start services
deploy() {
    print_status "Starting Chart Paper deployment..."
    
    # Navigate to devops directory
    cd "$(dirname "$0")"
    
    # Stop existing containers
    print_status "Stopping existing containers..."
    docker-compose down --remove-orphans
    
    # Build and start services
    print_status "Building and starting services..."
    docker-compose up --build -d
    
    # Wait for services to be healthy
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Check if services are running
    if docker-compose ps | grep -q "Up"; then
        print_success "Chart Paper deployed successfully!"
        echo ""
        echo "🚀 Chart Paper is now running:"
        echo "   Frontend: http://localhost:3000"
        echo "   Backend API: http://localhost:8000"
        echo ""
        echo "📊 To view logs:"
        echo "   docker-compose logs -f"
        echo ""
        echo "🛑 To stop:"
        echo "   docker-compose down"
    else
        print_error "Deployment failed. Check logs with: docker-compose logs"
        exit 1
    fi
}

# Function to show logs
logs() {
    cd "$(dirname "$0")"
    docker-compose logs -f
}

# Function to stop services
stop() {
    cd "$(dirname "$0")"
    print_status "Stopping Chart Paper..."
    docker-compose down
    print_success "Chart Paper stopped"
}

# Function to restart services
restart() {
    cd "$(dirname "$0")"
    print_status "Restarting Chart Paper..."
    docker-compose restart
    print_success "Chart Paper restarted"
}

# Function to show status
status() {
    cd "$(dirname "$0")"
    docker-compose ps
}

# Function to clean up
cleanup() {
    cd "$(dirname "$0")"
    print_warning "This will remove all containers, images, and volumes. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_status "Cleaning up Chart Paper..."
        docker-compose down --volumes --remove-orphans
        docker system prune -f
        print_success "Cleanup completed"
    else
        print_status "Cleanup cancelled"
    fi
}

# Main script logic
case "${1:-deploy}" in
    "deploy")
        check_docker
        deploy
        ;;
    "logs")
        logs
        ;;
    "stop")
        stop
        ;;
    "restart")
        restart
        ;;
    "status")
        status
        ;;
    "cleanup")
        cleanup
        ;;
    "help"|"-h"|"--help")
        echo "Chart Paper Deployment Script"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  deploy    Build and deploy Chart Paper (default)"
        echo "  logs      Show service logs"
        echo "  stop      Stop all services"
        echo "  restart   Restart all services"
        echo "  status    Show service status"
        echo "  cleanup   Remove all containers, images, and volumes"
        echo "  help      Show this help message"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
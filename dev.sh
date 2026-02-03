#!/usr/bin/env bash

set -e

COMPOSE="docker compose"
PROJECT_NAME="myapp"

usage() {
  echo "Usage: $0 {up|down|clean|restart|rebuild|logs}"
  echo ""
  echo "Commands:"
  echo "  up        Build and start containers"
  echo "  down      Stop containers"
  echo "  clean     Stop containers and remove volumes (DB, MinIO)"
  echo "  restart   Restart containers (rebuild images)"
  echo "  rebuild   Force rebuild all images without cache"
  echo "  logs      Show logs"
}

case "$1" in
  up)
    $COMPOSE -p $PROJECT_NAME up --build -d
    ;;
  down)
    $COMPOSE -p $PROJECT_NAME down
    ;;
  clean)
    $COMPOSE -p $PROJECT_NAME down -v
    ;;
  restart)
    echo "Stopping containers..."
    $COMPOSE -p $PROJECT_NAME down
    echo "Removing old images..."
    docker rmi -f $(docker images -q "${PROJECT_NAME}*" 2>/dev/null) 2>/dev/null || true
    echo "Building and starting containers..."
    $COMPOSE -p $PROJECT_NAME up --build -d
    ;;
  rebuild)
    echo "Stopping containers..."
    $COMPOSE -p $PROJECT_NAME down
    echo "Removing old images..."
    docker rmi -f $(docker images -q "${PROJECT_NAME}*" 2>/dev/null) 2>/dev/null || true
    echo "Building without cache..."
    $COMPOSE -p $PROJECT_NAME build --no-cache
    echo "Starting containers..."
    $COMPOSE -p $PROJECT_NAME up -d
    ;;
  logs)
    $COMPOSE -p $PROJECT_NAME logs -f
    ;;
  *)
    usage
    exit 1
    ;;
esac

# Variables
SRC_DIR := $(shell realpath "$$(dirname "$(CURDIR)")")
MAIN_FILE_PATH := example/main.go
SERVER_DEBUG_PORT := 40000
DOCKER_COMPOSE_PATH ?= docker-compose.yml

.PHONY: default help docker-up docker-down docker-restart docker-logs docker-ps

# Default target
default: help

# Help information
help:
	@echo "Available targets:"
	@echo "  docker-up       - Start Docker containers"
	@echo "  docker-down     - Stop Docker containers"
	@echo "  docker-restart  - Restart Docker containers"
	@echo "  docker-logs     - View logs from containers"
	@echo "  docker-ps       - List running containers"

# Docker commands
docker-up:
	cd $(SRC_DIR)/example/.docker && \
	SRC_DIR='$(SRC_DIR)' \
	MAIN_FILE_PATH='$(MAIN_FILE_PATH)' \
	SERVER_DEBUG_PORT='$(SERVER_DEBUG_PORT)' \
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d

docker-down:
	cd $(SRC_DIR)/example/.docker && \
	SRC_DIR='$(SRC_DIR)' \
	MAIN_FILE_PATH='$(MAIN_FILE_PATH)' \
	SERVER_DEBUG_PORT='$(SERVER_DEBUG_PORT)' \
	docker-compose -f $(DOCKER_COMPOSE_PATH) down

docker-restart: docker-down docker-up

docker-logs:
	cd $(SRC_DIR)/example/.docker && \
	SRC_DIR='$(SRC_DIR)' \
	MAIN_FILE_PATH='$(MAIN_FILE_PATH)' \
	SERVER_DEBUG_PORT='$(SERVER_DEBUG_PORT)' \
	docker-compose -f $(DOCKER_COMPOSE_PATH) logs -f

docker-ps:
	cd $(SRC_DIR)/example/.docker && \
	SRC_DIR='$(SRC_DIR)' \
	MAIN_FILE_PATH='$(MAIN_FILE_PATH)' \
	SERVER_DEBUG_PORT='$(SERVER_DEBUG_PORT)' \
	docker-compose -f $(DOCKER_COMPOSE_PATH) ps

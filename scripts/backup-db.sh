#!/bin/bash

# Database Backup Script for Dysgair Project
# Creates a timestamped backup of the MySQL database

set -e  # Exit on error

# Configuration
CONTAINER_NAME="dysgair-db-1"
DB_NAME="dysgair"
DB_USER="dysgair"
DB_PASSWORD=""
BACKUP_DIR="./backups"

# Generate timestamp for backup filename
TIMESTAMP=$(date +"%Y-%m-%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "  Dysgair Database Backup"
echo "========================================="
echo ""

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${RED}Error: Container '${CONTAINER_NAME}' is not running.${NC}"
    echo "Please start the container with: docker-compose up -d"
    exit 1
fi

# Check if backup directory exists
if [ ! -d "$BACKUP_DIR" ]; then
    echo -e "${YELLOW}Creating backup directory: ${BACKUP_DIR}${NC}"
    mkdir -p "$BACKUP_DIR"
fi

# Perform backup
echo -e "${YELLOW}Starting backup...${NC}"
echo "Database: ${DB_NAME}"
echo "Container: ${CONTAINER_NAME}"
echo "Output: ${BACKUP_FILE}"
echo ""

docker exec ${CONTAINER_NAME} mysqldump \
    -u ${DB_USER} \
    -p${DB_PASSWORD} \
    --no-create-info \
    --skip-triggers \
    --skip-lock-tables \
    --default-character-set=utf8mb4 \
    ${DB_NAME} > "${BACKUP_FILE}"

# Check if backup was successful
if [ $? -eq 0 ]; then
    BACKUP_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    echo -e "${GREEN}Backup completed successfully!${NC}"
    echo ""
    echo "Backup file: ${BACKUP_FILE}"
    echo "Size: ${BACKUP_SIZE}"
    echo ""

    # Count total backups
    BACKUP_COUNT=$(ls -1 ${BACKUP_DIR}/*.sql 2>/dev/null | wc -l)
    echo "Total backups in ${BACKUP_DIR}: ${BACKUP_COUNT}"
else
    echo -e "${RED}Backup failed!${NC}"
    exit 1
fi

echo ""
echo "========================================="

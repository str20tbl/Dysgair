#!/bin/bash

# Database Restore Script for Dysgair Project
# Restores the MySQL database from a backup file

set -e  # Exit on error

# Configuration
CONTAINER_NAME="dysgair-db-1"
DB_NAME="dysgair"
DB_USER="dysgair"
DB_PASSWORD="cjJaLbuCXGf9XJn94h9S3bes"
BACKUP_DIR="./backups"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "  Dysgair Database Restore"
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
    echo -e "${RED}Error: Backup directory '${BACKUP_DIR}' does not exist.${NC}"
    exit 1
fi

# List available backups
BACKUPS=($(ls -1t ${BACKUP_DIR}/*.sql 2>/dev/null))
BACKUP_COUNT=${#BACKUPS[@]}

if [ $BACKUP_COUNT -eq 0 ]; then
    echo -e "${RED}No backup files found in ${BACKUP_DIR}${NC}"
    echo "Please run ./backup-db.sh to create a backup first."
    exit 1
fi

echo -e "${BLUE}Available backups:${NC}"
echo ""

for i in "${!BACKUPS[@]}"; do
    BACKUP_FILE="${BACKUPS[$i]}"
    FILENAME=$(basename "$BACKUP_FILE")
    SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    DATE=$(stat -c %y "$BACKUP_FILE" | cut -d' ' -f1,2 | cut -d'.' -f1)
    printf "%2d) %-40s  %8s  %s\n" $((i+1)) "$FILENAME" "$SIZE" "$DATE"
done

echo ""
echo -e "${YELLOW}Enter the number of the backup to restore (or 'q' to quit):${NC}"
read -r CHOICE

# Check if user wants to quit
if [ "$CHOICE" = "q" ] || [ "$CHOICE" = "Q" ]; then
    echo "Restore cancelled."
    exit 0
fi

# Validate input
if ! [[ "$CHOICE" =~ ^[0-9]+$ ]] || [ "$CHOICE" -lt 1 ] || [ "$CHOICE" -gt $BACKUP_COUNT ]; then
    echo -e "${RED}Invalid selection.${NC}"
    exit 1
fi

# Get selected backup
SELECTED_BACKUP="${BACKUPS[$((CHOICE-1))]}"
SELECTED_FILENAME=$(basename "$SELECTED_BACKUP")

echo ""
echo -e "${YELLOW}WARNING: This will replace all data in the '${DB_NAME}' database!${NC}"
echo "Selected backup: ${SELECTED_FILENAME}"
echo ""
echo -e "${RED}Are you sure you want to continue? (yes/no):${NC}"
read -r CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Restore cancelled."
    exit 0
fi

# Perform restore
echo ""
echo -e "${YELLOW}Starting restore...${NC}"
echo "Database: ${DB_NAME}"
echo "Container: ${CONTAINER_NAME}"
echo "Backup file: ${SELECTED_FILENAME}"
echo ""

docker exec -i ${CONTAINER_NAME} mysql \
    -u ${DB_USER} \
    -p${DB_PASSWORD} \
    --default-character-set=utf8mb4 \
    ${DB_NAME} < "${SELECTED_BACKUP}"

# Check if restore was successful
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}Database restored successfully!${NC}"
    echo ""
else
    echo -e "${RED}Restore failed!${NC}"
    exit 1
fi

echo "========================================="

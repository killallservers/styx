#!/bin/bash
# PostgreSQL backup script
# Usage: docker exec allora-api-db /usr/local/bin/backup.sh

set -e

BACKUP_DIR="/var/lib/postgresql/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/postgres_${TIMESTAMP}.sql.gz"

# Create backups directory if it doesn't exist
mkdir -p "${BACKUP_DIR}"

echo "Starting backup at $(date)"

# Backup all databases
pg_dumpall -U allora | gzip > "${BACKUP_FILE}"

# Keep only last 7 days of backups
find "${BACKUP_DIR}" -name "postgres_*.sql.gz" -type f -mtime +7 -delete

echo "Backup completed: ${BACKUP_FILE}"
echo "Backup size: $(du -h ${BACKUP_FILE} | cut -f1)"

# List recent backups
echo "Recent backups:"
ls -lh "${BACKUP_DIR}" | tail -5

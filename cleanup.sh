#!/bin/sh

echo "[cleanup.sh] Starting cleanup at $(date)"

# Echo row count before
echo "[cleanup.sh] Total logs before:"
sqlite3 /data/logs.db "SELECT COUNT(*) FROM log;"

# Count logs older than 30 days
old_logs_count=$(sqlite3 /data/logs.db "SELECT COUNT(*) FROM log WHERE timestamp < datetime('now', '-30 days');")
echo "[cleanup.sh] Logs older than 30 days: $old_logs_count"

# Delete logs older than 30 days where severity_text is INFO
sqlite3 /data/logs.db <<EOF
DELETE FROM log
WHERE severity_text = 'INFO'
  AND timestamp < datetime('now', '-30 days');
EOF

# Echo row count after
echo "[cleanup.sh] Total logs after:"
sqlite3 /data/logs.db "SELECT COUNT(*) FROM log;"

echo "[cleanup.sh] Cleanup finished at $(date)"

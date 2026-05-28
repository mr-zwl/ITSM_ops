#!/bin/bash
set -e

ASSET_ID=""
ENDPOINT=""
INTERVAL=30

while [[ $# -gt 0 ]]; do
  case $1 in
    --asset-id) ASSET_ID="$2"; shift 2 ;;
    --endpoint) ENDPOINT="$2"; shift 2 ;;
    --interval) INTERVAL="$2"; shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [ -z "$ASSET_ID" ]; then echo "ERROR: --asset-id is required"; exit 1; fi
if [ -z "$ENDPOINT" ]; then ENDPOINT="http://CHANGE_ME_SERVER_IP:8080/api/v1/collect"; fi

echo "=== ITSM Ops Collector Installer (Linux) ==="
echo "Asset ID: $ASSET_ID"
echo "Endpoint: $ENDPOINT"

# Download
echo "Downloading collector..."
wget -q -O /usr/local/bin/itsm-collector http://CHANGE_ME_SERVER_IP:3000/downloads/itsm-collector-linux-amd64
chmod +x /usr/local/bin/itsm-collector

# Create systemd service
cat > /etc/systemd/system/itsm-collector.service << EOF
[Unit]
Description=ITSM Ops Collector
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/itsm-collector
Restart=always
RestartSec=5
Environment=COLLECTOR_ENDPOINT=$ENDPOINT
Environment=COLLECTOR_ASSET_ID=$ASSET_ID
Environment=COLLECTOR_INTERVAL=$INTERVAL
Environment=COLLECTOR_MODE=auto

[Install]
WantedBy=multi-user.target
EOF

# Start
systemctl daemon-reload
systemctl enable itsm-collector
systemctl start itsm-collector

echo "=== Installation complete! ==="
systemctl status itsm-collector --no-pager | head -5

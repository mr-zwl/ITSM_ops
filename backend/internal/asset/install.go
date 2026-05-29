package asset

import (
	"fmt"
	"net/http"
)

type InstallGuide struct {
	AssetID   uint64            `json:"asset_id"`
	AssetName string            `json:"asset_name"`
	OsType    string            `json:"os_type"`
	Endpoint  string            `json:"endpoint"`
	Commands  map[string]string `json:"commands"`
}

func installGuide(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	a, err := repo.GetAsset(r.Context(), id)
	if err != nil || a == nil {
		respondErr(w, http.StatusNotFound, "asset not found")
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	baseURL := scheme + "://" + host

	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		endpoint = baseURL + "/api/v1/collect"
	}

	osType := a.OsType
	if osType == "" {
		osType = "linux"
	}

	commands := map[string]string{}

	commands["linux"] = fmt.Sprintf(
		"curl -sL %s/downloads/install.sh | bash -s -- --asset-id=%d --endpoint=%s",
		baseURL, id, endpoint,
	)
	commands["windows"] = fmt.Sprintf(
		"Invoke-WebRequest -Uri %s/downloads/install.ps1 -OutFile install.ps1; ./install.ps1 -AssetId %d -Endpoint %s",
		baseURL, id, endpoint,
	)
	commands["linux_manual"] = fmt.Sprintf(`# 1. Download collector
wget -O /usr/local/bin/itsm-collector %s/downloads/itsm-collector-linux-amd64
chmod +x /usr/local/bin/itsm-collector

# 2. Create systemd service
cat > /etc/systemd/system/itsm-collector.service << 'EOF'
[Unit]
Description=ITSM Ops Collector
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/itsm-collector
Restart=always
RestartSec=5
Environment=COLLECTOR_ENDPOINT=%s
Environment=COLLECTOR_ASSET_ID=%d
Environment=COLLECTOR_INTERVAL=30
Environment=COLLECTOR_MODE=auto

[Install]
WantedBy=multi-user.target
EOF

# 3. Start
systemctl daemon-reload
systemctl enable itsm-collector
systemctl start itsm-collector`, baseURL, endpoint, id)

	commands["windows_manual"] = fmt.Sprintf(`# 1. Download
mkdir C:\ITSM-Agent -Force
Invoke-WebRequest -Uri %s/downloads/itsm-collector.exe -OutFile C:\ITSM-Agent\itsm-collector.exe

# 2. Create start.bat
Set-Content C:\ITSM-Agent\start.bat "@echo off
set COLLECTOR_ENDPOINT=%s
set COLLECTOR_ASSET_ID=%d
set COLLECTOR_INTERVAL=30
set COLLECTOR_MODE=auto
C:\ITSM-Agent\itsm-collector.exe"

# 3. Schedule
schtasks /create /tn ITSM-Collector /tr C:\ITSM-Agent\start.bat /sc onstart /ru SYSTEM /rl HIGHEST /f
schtasks /run /tn ITSM-Collector`, baseURL, endpoint, id)

	guide := InstallGuide{
		AssetID:   id,
		AssetName: a.Name,
		OsType:    osType,
		Endpoint:  endpoint,
		Commands:  commands,
	}

	respondOK(w, guide)
}

func registerInstallRoute(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/assets/{id}/install", installGuide)
}

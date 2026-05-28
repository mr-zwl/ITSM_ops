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

	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		endpoint = "http://CHANGE_ME_SERVER_IP:8080/api/v1/collect"
	}

	osType := a.OsType
	if osType == "" {
		osType = "linux"
	}

	commands := map[string]string{}

	commands["linux"] = fmt.Sprintf(
		"curl -sL http://CHANGE_ME_SERVER_IP:3000/downloads/install.sh | bash -s -- --asset-id=%d --endpoint=%s",
		id, endpoint,
	)
	commands["windows"] = fmt.Sprintf(
		"Invoke-WebRequest -Uri http://CHANGE_ME_SERVER_IP:3000/downloads/install.ps1 -OutFile install.ps1; ./install.ps1 -AssetId %d -Endpoint %s",
		id, endpoint,
	)
	commands["linux_manual"] = fmt.Sprintf(
		"# 1. Download collector\nwget -O /usr/local/bin/itsm-collector http://CHANGE_ME_SERVER_IP:3000/downloads/itsm-collector-linux-amd64\nchmod +x /usr/local/bin/itsm-collector\n\n# 2. Create systemd service\ncat > /etc/systemd/system/itsm-collector.service << EOF\n[Unit]\nDescription=ITSM Ops Collector\nAfter=network.target\n\n[Service]\nType=simple\nExecStart=/usr/local/bin/itsm-collector\nRestart=always\nRestartSec=5\nEnvironment=COLLECTOR_ENDPOINT=%s\nEnvironment=COLLECTOR_ASSET_ID=%d\nEnvironment=COLLECTOR_INTERVAL=30\nEnvironment=COLLECTOR_MODE=auto\n\n[Install]\nWantedBy=multi-user.target\nEOF\n\n# 3. Start\nsystemctl daemon-reload\nsystemctl enable itsm-collector\nsystemctl start itsm-collector",
		endpoint, id,
	)
	commands["windows_manual"] = fmt.Sprintf(
		"# 1. Download\nmkdir C:\\ITSM-Agent -Force\nInvoke-WebRequest -Uri http://CHANGE_ME_SERVER_IP:3000/downloads/itsm-collector.exe -OutFile C:\\ITSM-Agent\\itsm-collector.exe\n\n# 2. Create start.bat\nSet-Content C:\\ITSM-Agent\\start.bat \"@echo off\nset COLLECTOR_ENDPOINT=%s\nset COLLECTOR_ASSET_ID=%d\nset COLLECTOR_INTERVAL=30\nset COLLECTOR_MODE=auto\nC:\\ITSM-Agent\\itsm-collector.exe\"\n\n# 3. Schedule\nschtasks /create /tn ITSM-Collector /tr C:\\ITSM-Agent\\start.bat /sc onstart /ru SYSTEM /rl HIGHEST /f\nschtasks /run /tn ITSM-Collector",
		endpoint, id,
	)

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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type MetricPoint struct {
	AssetID    uint64  `json:"asset_id"`
	MetricCode string  `json:"metric_code"`
	Value      float64 `json:"value"`
}

type IngestPayload struct {
	Points []MetricPoint `json:"points"`
}

func main() {
	endpoint := envStr("COLLECTOR_ENDPOINT", "http://127.0.0.1:8080/api/v1/collect")
	token := envStr("COLLECTOR_TOKEN", "")
	assetID := envUint64("COLLECTOR_ASSET_ID", 1)
	interval := envInt("COLLECTOR_INTERVAL", 30)
	mode := envStr("COLLECTOR_MODE", "auto")

	fmt.Printf("collector starting: endpoint=%s asset_id=%d interval=%ds mode=%s\n", endpoint, assetID, interval, mode)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	collect(endpoint, token, assetID, mode)

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			collect(endpoint, token, assetID, mode)
		case <-stopCh:
			fmt.Println("collector stopped")
			return
		}
	}
}

func collect(endpoint, token string, assetID uint64, mode string) {
	var points []MetricPoint

	if mode == "mock" || runtime.GOOS != "linux" {
		points = mockMetrics(assetID)
	} else {
		points = realMetrics(assetID)
	}

	payload := IngestPayload{Points: points}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "request error: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "send error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("collected %d points -> %d\n", len(points), resp.StatusCode)
}

func realMetrics(assetID uint64) []MetricPoint {
	points := []MetricPoint{}

	if cpu, err := getCPUUsage(); err == nil {
		points = append(points, MetricPoint{assetID, "cpu_usage", cpu})
	}
	if mem, err := getMemUsage(); err == nil {
		points = append(points, MetricPoint{assetID, "mem_usage", mem})
	}
	if disk, err := getDiskUsage(); err == nil {
		points = append(points, MetricPoint{assetID, "disk_usage", disk})
	}
	if load, err := getLoadAvg(); err == nil {
		points = append(points, MetricPoint{assetID, "load_avg_1m", load})
	}

	return points
}

func mockMetrics(assetID uint64) []MetricPoint {
	return []MetricPoint{
		{assetID, "cpu_usage", 20 + rand.Float64()*60},
		{assetID, "mem_usage", 30 + rand.Float64()*50},
		{assetID, "disk_usage", 20 + rand.Float64()*40},
		{assetID, "load_avg_1m", rand.Float64() * 4},
		{assetID, "net_in_bytes", rand.Float64() * 1e8},
		{assetID, "net_out_bytes", rand.Float64() * 5e7},
	}
}

func getCPUUsage() (float64, error) {
	out, err := exec.Command("sh", "-c",
		`top -bn1 | grep '%Cpu' | awk '{print $2+$4}'`).Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

func getMemUsage() (float64, error) {
	out, err := exec.Command("sh", "-c",
		`free | awk '/Mem:/{printf("%.1f", $3/$2*100)}'`).Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

func getDiskUsage() (float64, error) {
	out, err := exec.Command("sh", "-c",
		`df / | awk 'NR==2{gsub(/%/,"",$5); print $5}'`).Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

func getLoadAvg() (float64, error) {
	out, err := exec.Command("sh", "-c",
		`cat /proc/loadavg | awk '{print $1}'`).Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envUint64(key string, fallback uint64) uint64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

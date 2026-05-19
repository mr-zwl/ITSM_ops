package model

import "time"

type User struct {
	ID          uint64    `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Role struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type AssetType struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Asset struct {
	ID           uint64    `json:"id"`
	AssetTypeID  uint64    `json:"asset_type_id"`
	Name         string    `json:"name"`
	IP           string    `json:"ip"`
	SN           string    `json:"sn"`
	Manufacturer string    `json:"manufacturer"`
	Model        string    `json:"model"`
	Location     string    `json:"location"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MetricDefinition struct {
	ID          uint64    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Unit        string    `json:"unit"`
	DataType    string    `json:"data_type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type CollectTask struct {
	ID            uint64     `json:"id"`
	AssetID       uint64     `json:"asset_id"`
	Protocol      string     `json:"protocol"`
	IntervalSec   int        `json:"interval_sec"`
	Enabled       bool       `json:"enabled"`
	LastCollectAt *time.Time `json:"last_collect_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AlertRule struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	MetricCode  string    `json:"metric_code"`
	ConditionOp string    `json:"condition_op"`
	Threshold   float64   `json:"threshold"`
	Severity    string    `json:"severity"`
	Duration    int       `json:"duration"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AlertEvent struct {
	ID         uint64     `json:"id"`
	RuleID     uint64     `json:"rule_id"`
	AssetID    *uint64    `json:"asset_id"`
	Severity   string     `json:"severity"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	FiredAt    time.Time  `json:"fired_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
	AckedBy    *uint64    `json:"acked_by"`
	AckedAt    *time.Time `json:"acked_at"`
}

type NotifyChannel struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type NotifyRecord struct {
	ID        uint64     `json:"id"`
	ChannelID uint64     `json:"channel_id"`
	EventID   uint64     `json:"event_id"`
	Status    string     `json:"status"`
	SentAt    *time.Time `json:"sent_at"`
	ErrorMsg  string     `json:"error_msg"`
	CreatedAt time.Time  `json:"created_at"`
}

type AuditLog struct {
	ID        uint64    `json:"id"`
	UserID    *uint64   `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

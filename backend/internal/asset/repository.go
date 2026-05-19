package asset

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Asset struct {
	ID           uint64 `db:"id"            json:"id"`
	AssetTypeID  uint64 `db:"asset_type_id" json:"asset_type_id"`
	Name         string `db:"name"          json:"name"`
	IP           string `db:"ip"            json:"ip"`
	SN           string `db:"sn"            json:"sn"`
	Manufacturer string `db:"manufacturer"  json:"manufacturer"`
	Model        string `db:"model"         json:"model"`
	Location     string `db:"location"      json:"location"`
	Status       string `db:"status"        json:"status"`
	Tags         string `db:"tags"          json:"tags"`
	Extra        string `db:"extra"         json:"extra"`
	CreatedAt    string `db:"created_at"    json:"created_at"`
	UpdatedAt    string `db:"updated_at"    json:"updated_at"`
}

type AssetType struct {
	ID   uint64 `db:"id"   json:"id"`
	Code string `db:"code" json:"code"`
	Name string `db:"name" json:"name"`
}

type CreateAssetInput struct {
	AssetTypeID  uint64 `json:"asset_type_id"`
	Name         string `json:"name"`
	IP           string `json:"ip"`
	SN           string `json:"sn"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Location     string `json:"location"`
	Status       string `json:"status"`
}

type UpdateAssetInput struct {
	Name         *string `json:"name"`
	IP           *string `json:"ip"`
	SN           *string `json:"sn"`
	Manufacturer *string `json:"manufacturer"`
	Model        *string `json:"model"`
	Location     *string `json:"location"`
	Status       *string `json:"status"`
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListAssets(ctx context.Context) ([]Asset, error) {
	var assets []Asset
	err := r.db.SelectContext(ctx, &assets, "SELECT * FROM assets ORDER BY id DESC")
	return assets, err
}

func (r *Repository) GetAsset(ctx context.Context, id uint64) (*Asset, error) {
	var a Asset
	err := r.db.GetContext(ctx, &a, "SELECT * FROM assets WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) CreateAsset(ctx context.Context, in CreateAssetInput) (uint64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO assets (asset_type_id, name, ip, sn, manufacturer, model, location, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		in.AssetTypeID, in.Name, in.IP, in.SN, in.Manufacturer, in.Model, in.Location, in.Status)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

func (r *Repository) UpdateAsset(ctx context.Context, id uint64, in UpdateAssetInput) error {
	a, err := r.GetAsset(ctx, id)
	if err != nil {
		return err
	}
	if a == nil {
		return sql.ErrNoRows
	}
	if in.Name != nil {
		a.Name = *in.Name
	}
	if in.IP != nil {
		a.IP = *in.IP
	}
	if in.SN != nil {
		a.SN = *in.SN
	}
	if in.Manufacturer != nil {
		a.Manufacturer = *in.Manufacturer
	}
	if in.Model != nil {
		a.Model = *in.Model
	}
	if in.Location != nil {
		a.Location = *in.Location
	}
	if in.Status != nil {
		a.Status = *in.Status
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE assets SET name=?, ip=?, sn=?, manufacturer=?, model=?, location=?, status=?, updated_at=datetime('now') WHERE id=?`,
		a.Name, a.IP, a.SN, a.Manufacturer, a.Model, a.Location, a.Status, id)
	return err
}

func (r *Repository) DeleteAsset(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM assets WHERE id = ?", id)
	return err
}

func (r *Repository) ListTypes(ctx context.Context) ([]AssetType, error) {
	var types []AssetType
	err := r.db.SelectContext(ctx, &types, "SELECT * FROM asset_types ORDER BY id")
	return types, err
}

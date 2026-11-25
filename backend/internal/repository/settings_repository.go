package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"ecommerce-backend/internal/models"
)

// SettingsRepository reads and writes the current store's settings row
// (row-level security limits store_settings to that one row).
type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository { return &SettingsRepository{db: db} }

func (r *SettingsRepository) Get(ctx context.Context) (models.StoreSettings, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx, `SELECT data FROM store_settings`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return models.StoreSettings{}, nil // a new store that never saved settings: the frontend's defaults apply
	}
	if err != nil {
		return models.StoreSettings{}, err
	}
	var s models.StoreSettings
	if err := json.Unmarshal(raw, &s); err != nil {
		return models.StoreSettings{}, err
	}
	return s, nil
}

func (r *SettingsRepository) Save(ctx context.Context, s models.StoreSettings) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO store_settings (data, updated_at) VALUES ($1, NOW())
		ON CONFLICT (store_id) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()`, raw)
	return err
}

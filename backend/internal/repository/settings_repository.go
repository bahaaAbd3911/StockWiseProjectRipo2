package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"ecommerce-backend/internal/models"
)

// SettingsRepository reads and writes the single store_settings row.
type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository { return &SettingsRepository{db: db} }

func (r *SettingsRepository) Get(ctx context.Context) (models.StoreSettings, error) {
	var raw []byte
	if err := r.db.QueryRowContext(ctx, `SELECT data FROM store_settings WHERE id = 1`).Scan(&raw); err != nil {
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
		INSERT INTO store_settings (id, data, updated_at) VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()`, raw)
	return err
}

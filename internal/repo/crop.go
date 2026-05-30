package repo

import (
	"database/sql"
	"imagemanager/internal/model"
)

type CropRepo struct{ db *sql.DB }

func NewCropRepo(db *sql.DB) *CropRepo { return &CropRepo{db: db} }

func (r *CropRepo) ForImage(imageID string) ([]model.Crop, error) {
	rows, err := r.db.Query("SELECT id, image_id, preset_id, variant_id, crop_x, crop_y, crop_w, crop_h, generated_at, updated_at FROM crops WHERE image_id = ?", imageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var crops []model.Crop
	for rows.Next() {
		var c model.Crop
		if err := rows.Scan(&c.ID, &c.ImageID, &c.PresetID, &c.VariantID, &c.CropX, &c.CropY, &c.CropW, &c.CropH, &c.GeneratedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		crops = append(crops, c)
	}
	return crops, rows.Err()
}

func (r *CropRepo) Find(imageID, presetID, variantID string) (*model.Crop, error) {
	row := r.db.QueryRow(`
		SELECT id, image_id, preset_id, variant_id, crop_x, crop_y, crop_w, crop_h, generated_at, updated_at
		FROM crops WHERE image_id = ? AND preset_id = ? AND variant_id = ?`,
		imageID, presetID, variantID,
	)
	var c model.Crop
	err := row.Scan(&c.ID, &c.ImageID, &c.PresetID, &c.VariantID, &c.CropX, &c.CropY, &c.CropW, &c.CropH, &c.GeneratedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CropRepo) Upsert(imageID, presetID, variantID string, x, y, w, h int) error {
	_, err := r.db.Exec(`
		INSERT INTO crops (image_id, preset_id, variant_id, crop_x, crop_y, crop_w, crop_h, generated_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NULL, datetime('now'))
		ON CONFLICT(image_id, preset_id, variant_id) DO UPDATE SET
			crop_x = excluded.crop_x,
			crop_y = excluded.crop_y,
			crop_w = excluded.crop_w,
			crop_h = excluded.crop_h,
			generated_at = NULL,
			updated_at = datetime('now')`,
		imageID, presetID, variantID, x, y, w, h,
	)
	return err
}

func (r *CropRepo) MarkGenerated(imageID, presetID, variantID string) error {
	_, err := r.db.Exec(`
		UPDATE crops SET generated_at = datetime('now')
		WHERE image_id = ? AND preset_id = ? AND variant_id = ?`,
		imageID, presetID, variantID,
	)
	return err
}

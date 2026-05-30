package repo

import (
	"database/sql"
	"imagemanager/internal/model"
)

type ComponentRepo struct{ db *sql.DB }

func NewComponentRepo(db *sql.DB) *ComponentRepo { return &ComponentRepo{db: db} }

func (r *ComponentRepo) ForComponent(componentID string) ([]model.ComponentImage, error) {
	rows, err := r.db.Query(`
		SELECT ci.id, ci.component_id, ci.is_active, ci.assigned_at,
		       i.id, i.original_name, i.name, i.filename, i.width, i.height, i.file_size, i.created_at
		FROM component_images ci
		JOIN images i ON i.id = ci.image_id
		WHERE ci.component_id = ?
		ORDER BY ci.assigned_at DESC`, componentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanComponentImages(rows)
}

func (r *ComponentRepo) ActiveImage(componentID string) (*model.ComponentImage, error) {
	rows, err := r.db.Query(`
		SELECT ci.id, ci.component_id, ci.is_active, ci.assigned_at,
		       i.id, i.original_name, i.name, i.filename, i.width, i.height, i.file_size, i.created_at
		FROM component_images ci
		JOIN images i ON i.id = ci.image_id
		WHERE ci.component_id = ? AND ci.is_active = 1
		LIMIT 1`, componentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanComponentImages(rows)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *ComponentRepo) Assign(componentID, imageID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE component_images SET is_active = 0 WHERE component_id = ?", componentID); err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO component_images (component_id, image_id, is_active) VALUES (?, ?, 1)", componentID, imageID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ComponentRepo) SetActive(assignmentID int64) error {
	row := r.db.QueryRow("SELECT component_id FROM component_images WHERE id = ?", assignmentID)
	var componentID string
	if err := row.Scan(&componentID); err != nil {
		return nil // not found, no-op
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE component_images SET is_active = 0 WHERE component_id = ?", componentID); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE component_images SET is_active = 1 WHERE id = ?", assignmentID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ComponentRepo) RemoveAssignment(assignmentID int64) error {
	row := r.db.QueryRow("SELECT component_id, is_active FROM component_images WHERE id = ?", assignmentID)
	var componentID string
	var isActive int
	if err := row.Scan(&componentID, &isActive); err != nil {
		return nil // not found, no-op
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM component_images WHERE id = ?", assignmentID); err != nil {
		return err
	}

	if isActive == 1 {
		// Promote next most recent
		if _, err := tx.Exec(`
			UPDATE component_images SET is_active = 1
			WHERE component_id = ?
			ORDER BY assigned_at DESC
			LIMIT 1`, componentID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ComponentRepo) AllForImage(imageID string) ([]model.ComponentImage, error) {
	rows, err := r.db.Query(`
		SELECT ci.id, ci.component_id, ci.is_active, ci.assigned_at,
		       i.id, i.original_name, i.name, i.filename, i.width, i.height, i.file_size, i.created_at
		FROM component_images ci
		JOIN images i ON i.id = ci.image_id
		WHERE ci.image_id = ?`, imageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanComponentImages(rows)
}

func (r *ComponentRepo) GenerationStatusRaw(componentID string, presetIDs []string, variantCounts map[string][]string) (int, int, error) {
	row := r.db.QueryRow(`
		SELECT ci.image_id FROM component_images ci
		WHERE ci.component_id = ? AND ci.is_active = 1
		LIMIT 1`, componentID)
	var imageID string
	if err := row.Scan(&imageID); err != nil {
		return 0, 0, nil
	}

	total, generated := 0, 0
	for _, presetID := range presetIDs {
		variants := variantCounts[presetID]
		for _, variantID := range variants {
			total++
			row := r.db.QueryRow(`
				SELECT generated_at FROM crops
				WHERE image_id = ? AND preset_id = ? AND variant_id = ?`,
				imageID, presetID, variantID)
			var genAt sql.NullString
			if err := row.Scan(&genAt); err == nil && genAt.Valid {
				generated++
			}
		}
	}
	return total, generated, nil
}

func scanComponentImages(rows *sql.Rows) ([]model.ComponentImage, error) {
	var items []model.ComponentImage
	for rows.Next() {
		var ci model.ComponentImage
		var isActive int
		var name sql.NullString
		if err := rows.Scan(
			&ci.AssignmentID, &ci.ComponentID, &isActive, &ci.AssignedAt,
			&ci.ImageID, &ci.OriginalName, &name, &ci.Filename,
			&ci.Width, &ci.Height, &ci.FileSize, &ci.CreatedAt,
		); err != nil {
			return nil, err
		}
		ci.IsActive = isActive == 1
		ci.Name = name.String
		if ci.Name == "" {
			ci.Name = ci.OriginalName
		}
		items = append(items, ci)
	}
	return items, rows.Err()
}

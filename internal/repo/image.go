package repo

import (
	"database/sql"
	"fmt"
	"imagemanager/internal/model"
	"strings"
)

type ImageRepo struct{ db *sql.DB }

func NewImageRepo(db *sql.DB) *ImageRepo { return &ImageRepo{db: db} }

func (r *ImageRepo) Insert(img model.Image) error {
	_, err := r.db.Exec(`
		INSERT INTO images (id, original_name, name, filename, mime_type, width, height, file_size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		img.ID, img.OriginalName, img.Name, img.Filename,
		img.MimeType, img.Width, img.Height, img.FileSize,
	)
	return err
}

func (r *ImageRepo) FindAll(search string, tagIDs []int64) ([]model.Image, error) {
	var args []any
	var where []string

	if search != "" {
		like := "%" + search + "%"
		where = append(where, "(LOWER(i.name) LIKE LOWER(?) OR LOWER(i.filename) LIKE LOWER(?))")
		args = append(args, like, like)
	}

	if len(tagIDs) > 0 {
		ph := strings.Repeat("?,", len(tagIDs))
		ph = ph[:len(ph)-1]
		where = append(where, fmt.Sprintf("i.id IN (SELECT image_id FROM image_tags WHERE tag_id IN (%s))", ph))
		for _, id := range tagIDs {
			args = append(args, id)
		}
	}

	whereSql := ""
	if len(where) > 0 {
		whereSql = "WHERE " + strings.Join(where, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT i.id, i.original_name, i.name, i.filename, i.mime_type,
		       i.width, i.height, i.file_size, i.created_at,
		       COUNT(CASE WHEN c.generated_at IS NOT NULL THEN 1 END) AS generated_count,
		       COUNT(c.id) AS crop_count
		FROM images i
		LEFT JOIN crops c ON c.image_id = i.id
		%s
		GROUP BY i.id
		ORDER BY i.created_at DESC`, whereSql)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []model.Image
	for rows.Next() {
		var img model.Image
		var name sql.NullString
		if err := rows.Scan(
			&img.ID, &img.OriginalName, &name, &img.Filename, &img.MimeType,
			&img.Width, &img.Height, &img.FileSize, &img.CreatedAt,
			&img.GeneratedCount, &img.CropCount,
		); err != nil {
			return nil, err
		}
		img.Name = name.String
		if img.Name == "" {
			img.Name = img.OriginalName
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

func (r *ImageRepo) FindByID(id string) (*model.Image, error) {
	row := r.db.QueryRow("SELECT id, original_name, name, filename, mime_type, width, height, file_size, created_at FROM images WHERE id = ?", id)
	return scanImage(row)
}

func (r *ImageRepo) FindByFilename(filename string) (*model.Image, error) {
	row := r.db.QueryRow("SELECT id, original_name, name, filename, mime_type, width, height, file_size, created_at FROM images WHERE filename = ?", filename)
	return scanImage(row)
}

func (r *ImageRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM images WHERE id = ?", id)
	return err
}

func scanImage(row *sql.Row) (*model.Image, error) {
	var img model.Image
	var name sql.NullString
	err := row.Scan(&img.ID, &img.OriginalName, &name, &img.Filename, &img.MimeType, &img.Width, &img.Height, &img.FileSize, &img.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	img.Name = name.String
	if img.Name == "" {
		img.Name = img.OriginalName
	}
	return &img, nil
}

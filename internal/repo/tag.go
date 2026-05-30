package repo

import (
	"database/sql"
	"fmt"
	"imagemanager/internal/model"
	"strings"
)

type TagRepo struct{ db *sql.DB }

func NewTagRepo(db *sql.DB) *TagRepo { return &TagRepo{db: db} }

func (r *TagRepo) All() ([]model.Tag, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.name, COUNT(it.image_id) AS image_count
		FROM tags t
		LEFT JOIN image_tags it ON it.tag_id = t.id
		GROUP BY t.id
		ORDER BY t.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.ImageCount); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *TagRepo) FindByName(name string) (*model.Tag, error) {
	row := r.db.QueryRow("SELECT id, name FROM tags WHERE LOWER(name) = LOWER(?)", strings.TrimSpace(name))
	var t model.Tag
	err := row.Scan(&t.ID, &t.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TagRepo) Create(name string) (int64, error) {
	res, err := r.db.Exec("INSERT INTO tags (name) VALUES (?)", strings.TrimSpace(name))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TagRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM tags WHERE id = ?", id)
	return err
}

func (r *TagRepo) ForImage(imageID string) ([]model.Tag, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.name
		FROM tags t
		JOIN image_tags it ON it.tag_id = t.id
		WHERE it.image_id = ?
		ORDER BY t.name ASC`, imageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *TagRepo) ForImages(imageIDs []string) (map[string][]model.Tag, error) {
	result := make(map[string][]model.Tag)
	if len(imageIDs) == 0 {
		return result, nil
	}

	ph := strings.Repeat("?,", len(imageIDs))
	ph = ph[:len(ph)-1]
	query := fmt.Sprintf(`
		SELECT it.image_id, t.id, t.name
		FROM image_tags it
		JOIN tags t ON t.id = it.tag_id
		WHERE it.image_id IN (%s)
		ORDER BY t.name ASC`, ph)

	args := make([]any, len(imageIDs))
	for i, id := range imageIDs {
		args[i] = id
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var imageID string
		var t model.Tag
		if err := rows.Scan(&imageID, &t.ID, &t.Name); err != nil {
			return nil, err
		}
		result[imageID] = append(result[imageID], t)
	}
	return result, rows.Err()
}

func (r *TagRepo) SetForImage(imageID string, tagIDs []int64) error {
	if _, err := r.db.Exec("DELETE FROM image_tags WHERE image_id = ?", imageID); err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if _, err := r.db.Exec("INSERT OR IGNORE INTO image_tags (image_id, tag_id) VALUES (?, ?)", imageID, tagID); err != nil {
			return err
		}
	}
	return nil
}

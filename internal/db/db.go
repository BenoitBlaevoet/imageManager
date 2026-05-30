package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(appRoot string) (*sql.DB, error) {
	dir := filepath.Join(appRoot, "data")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, "imagemanager.db"))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS images (
			id            TEXT PRIMARY KEY,
			original_name TEXT NOT NULL,
			name          TEXT,
			filename      TEXT NOT NULL UNIQUE,
			mime_type     TEXT NOT NULL,
			width         INTEGER NOT NULL,
			height        INTEGER NOT NULL,
			file_size     INTEGER NOT NULL,
			created_at    TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE TABLE IF NOT EXISTS crops (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			image_id     TEXT NOT NULL REFERENCES images(id) ON DELETE CASCADE,
			preset_id    TEXT NOT NULL,
			variant_id   TEXT NOT NULL,
			crop_x       INTEGER NOT NULL DEFAULT 0,
			crop_y       INTEGER NOT NULL DEFAULT 0,
			crop_w       INTEGER NOT NULL DEFAULT 0,
			crop_h       INTEGER NOT NULL DEFAULT 0,
			generated_at TEXT,
			updated_at   TEXT NOT NULL DEFAULT (datetime('now')),
			UNIQUE(image_id, preset_id, variant_id)
		);

		CREATE INDEX IF NOT EXISTS idx_crops_image ON crops(image_id);

		CREATE TABLE IF NOT EXISTS component_images (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			component_id TEXT NOT NULL,
			image_id     TEXT NOT NULL REFERENCES images(id) ON DELETE CASCADE,
			is_active    INTEGER NOT NULL DEFAULT 1,
			assigned_at  TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE INDEX IF NOT EXISTS idx_component_images_component ON component_images(component_id);
		CREATE INDEX IF NOT EXISTS idx_component_images_image ON component_images(image_id);

		CREATE TABLE IF NOT EXISTS tags (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS image_tags (
			image_id TEXT    NOT NULL REFERENCES images(id)  ON DELETE CASCADE,
			tag_id   INTEGER NOT NULL REFERENCES tags(id)    ON DELETE CASCADE,
			PRIMARY KEY (image_id, tag_id)
		);

		CREATE INDEX IF NOT EXISTS idx_image_tags_image ON image_tags(image_id);
		CREATE INDEX IF NOT EXISTS idx_image_tags_tag   ON image_tags(tag_id);
	`)
	if err != nil {
		return err
	}

	// Add name column to existing databases that predate this column
	if !columnExists(db, "images", "name") {
		if _, err := db.Exec("ALTER TABLE images ADD COLUMN name TEXT"); err != nil {
			return err
		}
	}

	_, err = db.Exec("UPDATE images SET name = original_name WHERE name IS NULL OR name = ''")
	return err
}

func columnExists(db *sql.DB, table, column string) bool {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			continue
		}
		if name == column {
			return true
		}
	}
	return false
}

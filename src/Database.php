<?php

namespace App;

use PDO;
use PDOException;

class Database
{
    private static ?PDO $instance = null;

    public static function get(): PDO
    {
        if (self::$instance === null) {
            $dbPath = dirname(__DIR__) . '/data/imagemanager.db';
            if (!is_dir(dirname($dbPath))) {
                mkdir(dirname($dbPath), 0755, true);
            }
            self::$instance = new PDO('sqlite:' . $dbPath);
            self::$instance->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
            self::$instance->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);
            self::$instance->exec('PRAGMA foreign_keys = ON');
            self::migrate(self::$instance);
        }

        return self::$instance;
    }

    private static function migrate(PDO $pdo): void
    {
        $pdo->exec("
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
        ");

        // Add name column to existing databases (silently skip if already present)
        try {
            $pdo->exec("ALTER TABLE images ADD COLUMN name TEXT");
        } catch (PDOException) {}

        $pdo->exec("UPDATE images SET name = original_name WHERE name IS NULL OR name = ''");
    }
}

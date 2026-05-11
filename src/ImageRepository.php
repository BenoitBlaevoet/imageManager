<?php

namespace App;

use PDO;

class ImageRepository
{
    public function __construct(private PDO $db) {}

    public function insert(array $data): void
    {
        $stmt = $this->db->prepare("
            INSERT INTO images (id, original_name, name, filename, mime_type, width, height, file_size)
            VALUES (:id, :original_name, :name, :filename, :mime_type, :width, :height, :file_size)
        ");
        $stmt->execute($data);
    }

    public function findAll(string $search = '', array $tagIds = []): array
    {
        $bindings = [];
        $where    = [];

        if ($search !== '') {
            $like    = '%' . $search . '%';
            $where[] = "(LOWER(i.name) LIKE LOWER(?) OR LOWER(i.filename) LIKE LOWER(?))";
            $bindings[] = $like;
            $bindings[] = $like;
        }

        if (!empty($tagIds)) {
            $ph      = implode(',', array_fill(0, count($tagIds), '?'));
            $where[] = "i.id IN (SELECT image_id FROM image_tags WHERE tag_id IN ($ph))";
            foreach ($tagIds as $id) {
                $bindings[] = (int) $id;
            }
        }

        $whereSql = $where ? 'WHERE ' . implode(' AND ', $where) : '';

        $stmt = $this->db->prepare("
            SELECT i.*,
                   COUNT(CASE WHEN c.generated_at IS NOT NULL THEN 1 END) AS generated_count,
                   COUNT(c.id) AS crop_count
            FROM images i
            LEFT JOIN crops c ON c.image_id = i.id
            $whereSql
            GROUP BY i.id
            ORDER BY i.created_at DESC
        ");
        $stmt->execute($bindings);
        return $stmt->fetchAll();
    }

    public function findById(string $id): ?array
    {
        $stmt = $this->db->prepare("SELECT * FROM images WHERE id = ?");
        $stmt->execute([$id]);
        $row = $stmt->fetch();
        return $row ?: null;
    }

    public function findByFilename(string $filename): ?array
    {
        $stmt = $this->db->prepare("SELECT * FROM images WHERE filename = ?");
        $stmt->execute([$filename]);
        $row = $stmt->fetch();
        return $row ?: null;
    }

    public function delete(string $id): void
    {
        $stmt = $this->db->prepare("DELETE FROM images WHERE id = ?");
        $stmt->execute([$id]);
    }
}

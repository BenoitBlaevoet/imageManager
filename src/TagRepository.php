<?php

namespace App;

use PDO;

class TagRepository
{
    public function __construct(private PDO $db) {}

    public function all(): array
    {
        return $this->db->query("
            SELECT t.id, t.name, COUNT(it.image_id) AS image_count
            FROM tags t
            LEFT JOIN image_tags it ON it.tag_id = t.id
            GROUP BY t.id
            ORDER BY t.name ASC
        ")->fetchAll();
    }

    public function findById(int $id): ?array
    {
        $stmt = $this->db->prepare("SELECT * FROM tags WHERE id = ?");
        $stmt->execute([$id]);
        return $stmt->fetch() ?: null;
    }

    public function findByName(string $name): ?array
    {
        $stmt = $this->db->prepare("SELECT * FROM tags WHERE LOWER(name) = LOWER(?)");
        $stmt->execute([trim($name)]);
        return $stmt->fetch() ?: null;
    }

    public function create(string $name): int
    {
        $stmt = $this->db->prepare("INSERT INTO tags (name) VALUES (?)");
        $stmt->execute([trim($name)]);
        return (int) $this->db->lastInsertId();
    }

    public function delete(int $id): void
    {
        $stmt = $this->db->prepare("DELETE FROM tags WHERE id = ?");
        $stmt->execute([$id]);
    }

    public function tagsForImage(string $imageId): array
    {
        $stmt = $this->db->prepare("
            SELECT t.id, t.name
            FROM tags t
            JOIN image_tags it ON it.tag_id = t.id
            WHERE it.image_id = ?
            ORDER BY t.name ASC
        ");
        $stmt->execute([$imageId]);
        return $stmt->fetchAll();
    }

    public function tagsForImages(array $imageIds): array
    {
        if (empty($imageIds)) {
            return [];
        }

        $placeholders = implode(',', array_fill(0, count($imageIds), '?'));
        $stmt = $this->db->prepare("
            SELECT it.image_id, t.id, t.name
            FROM image_tags it
            JOIN tags t ON t.id = it.tag_id
            WHERE it.image_id IN ($placeholders)
            ORDER BY t.name ASC
        ");
        $stmt->execute($imageIds);

        $result = [];
        foreach ($stmt->fetchAll() as $row) {
            $result[$row['image_id']][] = ['id' => (int) $row['id'], 'name' => $row['name']];
        }
        return $result;
    }

    public function setTagsForImage(string $imageId, array $tagIds): void
    {
        $this->db->prepare("DELETE FROM image_tags WHERE image_id = ?")->execute([$imageId]);

        if (empty($tagIds)) {
            return;
        }

        $stmt = $this->db->prepare("INSERT OR IGNORE INTO image_tags (image_id, tag_id) VALUES (?, ?)");
        foreach ($tagIds as $tagId) {
            $stmt->execute([$imageId, (int) $tagId]);
        }
    }
}

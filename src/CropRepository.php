<?php

namespace App;

use PDO;

class CropRepository
{
    public function __construct(private PDO $db) {}

    public function forImage(string $imageId): array
    {
        $stmt = $this->db->prepare("SELECT * FROM crops WHERE image_id = ?");
        $stmt->execute([$imageId]);
        return $stmt->fetchAll();
    }

    public function find(string $imageId, string $presetId, string $variantId): ?array
    {
        $stmt = $this->db->prepare("
            SELECT * FROM crops
            WHERE image_id = ? AND preset_id = ? AND variant_id = ?
        ");
        $stmt->execute([$imageId, $presetId, $variantId]);
        $row = $stmt->fetch();
        return $row ?: null;
    }

    public function upsert(string $imageId, string $presetId, string $variantId, int $x, int $y, int $w, int $h): void
    {
        $stmt = $this->db->prepare("
            INSERT INTO crops (image_id, preset_id, variant_id, crop_x, crop_y, crop_w, crop_h, generated_at, updated_at)
            VALUES (:image_id, :preset_id, :variant_id, :crop_x, :crop_y, :crop_w, :crop_h, NULL, datetime('now'))
            ON CONFLICT(image_id, preset_id, variant_id) DO UPDATE SET
                crop_x = excluded.crop_x,
                crop_y = excluded.crop_y,
                crop_w = excluded.crop_w,
                crop_h = excluded.crop_h,
                generated_at = NULL,
                updated_at = datetime('now')
        ");
        $stmt->execute([
            'image_id'  => $imageId,
            'preset_id' => $presetId,
            'variant_id'=> $variantId,
            'crop_x'    => $x,
            'crop_y'    => $y,
            'crop_w'    => $w,
            'crop_h'    => $h,
        ]);
    }

    public function markGenerated(string $imageId, string $presetId, string $variantId): void
    {
        $stmt = $this->db->prepare("
            UPDATE crops SET generated_at = datetime('now')
            WHERE image_id = ? AND preset_id = ? AND variant_id = ?
        ");
        $stmt->execute([$imageId, $presetId, $variantId]);
    }

    public function deleteForImage(string $imageId): void
    {
        $stmt = $this->db->prepare("DELETE FROM crops WHERE image_id = ?");
        $stmt->execute([$imageId]);
    }
}

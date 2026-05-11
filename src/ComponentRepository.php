<?php

namespace App;

use PDO;

class ComponentRepository
{
    public function __construct(private PDO $db) {}

    public function forComponent(string $componentId): array
    {
        $stmt = $this->db->prepare("
            SELECT ci.id as assignment_id, ci.component_id, ci.is_active, ci.assigned_at,
                   i.id, i.original_name, i.name, i.filename, i.width, i.height, i.file_size, i.created_at
            FROM component_images ci
            JOIN images i ON i.id = ci.image_id
            WHERE ci.component_id = ?
            ORDER BY ci.assigned_at DESC
        ");
        $stmt->execute([$componentId]);
        return $stmt->fetchAll();
    }

    public function activeImage(string $componentId): ?array
    {
        $stmt = $this->db->prepare("
            SELECT ci.id as assignment_id, ci.component_id, ci.is_active, ci.assigned_at,
                   i.id, i.original_name, i.name, i.filename, i.width, i.height, i.file_size, i.created_at
            FROM component_images ci
            JOIN images i ON i.id = ci.image_id
            WHERE ci.component_id = ? AND ci.is_active = 1
            LIMIT 1
        ");
        $stmt->execute([$componentId]);
        $row = $stmt->fetch();
        return $row ?: null;
    }

    public function assign(string $componentId, string $imageId): void
    {
        // Deactivate all existing assignments for this component
        $stmt = $this->db->prepare("
            UPDATE component_images SET is_active = 0 WHERE component_id = ?
        ");
        $stmt->execute([$componentId]);

        // Insert new assignment as active
        $stmt = $this->db->prepare("
            INSERT INTO component_images (component_id, image_id, is_active)
            VALUES (?, ?, 1)
        ");
        $stmt->execute([$componentId, $imageId]);
    }

    public function setActive(int $assignmentId): void
    {
        // Get the component_id for this assignment
        $stmt = $this->db->prepare("SELECT component_id FROM component_images WHERE id = ?");
        $stmt->execute([$assignmentId]);
        $row = $stmt->fetch();
        if (!$row) {
            return;
        }

        // Deactivate all for this component, then activate the target
        $this->db->prepare("UPDATE component_images SET is_active = 0 WHERE component_id = ?")
            ->execute([$row['component_id']]);

        $this->db->prepare("UPDATE component_images SET is_active = 1 WHERE id = ?")
            ->execute([$assignmentId]);
    }

    public function removeAssignment(int $assignmentId): void
    {
        // If removing the active one, promote the next most recent
        $stmt = $this->db->prepare("SELECT component_id, is_active FROM component_images WHERE id = ?");
        $stmt->execute([$assignmentId]);
        $row = $stmt->fetch();

        $this->db->prepare("DELETE FROM component_images WHERE id = ?")->execute([$assignmentId]);

        if ($row && $row['is_active']) {
            // Promote the next most recent assignment
            $stmt = $this->db->prepare("
                UPDATE component_images SET is_active = 1
                WHERE component_id = ?
                ORDER BY assigned_at DESC
                LIMIT 1
            ");
            $stmt->execute([$row['component_id']]);
        }
    }

    public function allAssignmentsForImage(string $imageId): array
    {
        $stmt = $this->db->prepare("SELECT * FROM component_images WHERE image_id = ?");
        $stmt->execute([$imageId]);
        return $stmt->fetchAll();
    }

    public function generationStatus(string $componentId, array $presets): array
    {
        $activeImage = $this->activeImage($componentId);
        if (!$activeImage) {
            return ['total' => 0, 'generated' => 0];
        }

        $imageId     = $activeImage['id'];
        $variantCount = 0;
        $genCount    = 0;

        foreach ($presets as $presetId) {
            $preset = PresetLoader::findPreset($presetId);
            if (!$preset) continue;
            foreach ($preset['variants'] as $variant) {
                $variantCount++;
                $stmt = $this->db->prepare("
                    SELECT generated_at FROM crops
                    WHERE image_id = ? AND preset_id = ? AND variant_id = ?
                ");
                $stmt->execute([$imageId, $presetId, $variant['id']]);
                $crop = $stmt->fetch();
                if ($crop && $crop['generated_at'] !== null) {
                    $genCount++;
                }
            }
        }

        return ['total' => $variantCount, 'generated' => $genCount];
    }
}

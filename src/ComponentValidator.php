<?php

namespace App;

class ComponentValidator
{
    public static function validateComponents(array $components): array
    {
        $errors = [];
        $ids    = [];

        foreach ($components as $i => $component) {
            $path = "components[$i]";

            if (empty($component['id']) || !is_string($component['id'])) {
                $errors[] = "$path.id is required";
            } elseif (!preg_match('/^[a-z0-9-]+$/', $component['id'])) {
                $errors[] = "$path.id must contain only lowercase letters, numbers, and hyphens";
            } elseif (in_array($component['id'], $ids, true)) {
                $errors[] = "$path.id '{$component['id']}' is already used";
            } else {
                $ids[] = $component['id'];
            }

            if (empty($component['label']) || !is_string($component['label'])) {
                $errors[] = "$path.label is required";
            }

            if (!isset($component['presets']) || !is_array($component['presets']) || count($component['presets']) === 0) {
                $errors[] = "$path must have at least one preset";
            } else {
                // Verify each preset ID exists
                foreach ($component['presets'] as $presetId) {
                    if (PresetLoader::findPreset($presetId) === null) {
                        $errors[] = "$path preset '$presetId' does not exist in presets.json";
                    }
                }
            }
        }

        return $errors;
    }

    public static function generateId(string $label): string
    {
        $id = strtolower($label);
        $id = preg_replace('/[^a-z0-9-]/', '-', $id);
        $id = preg_replace('/-+/', '-', $id);
        return trim($id, '-');
    }
}

<?php

namespace App;

class PresetValidator
{
    public static function validatePresets(array $presets): array
    {
        $errors = [];
        $presetIds = [];
        $index = 0;

        foreach ($presets as $preset) {
            $path = "presets[$index]";

            // Check structure
            if (!isset($preset['id']) || !is_string($preset['id']) || $preset['id'] === '') {
                $errors[] = "$path.id is required and must be a non-empty string";
            }

            if (!isset($preset['label']) || !is_string($preset['label']) || $preset['label'] === '') {
                $errors[] = "$path.label is required and must be a non-empty string";
            }

            // Validate ID format
            if (isset($preset['id']) && $preset['id'] !== '') {
                if (!preg_match('/^[a-z0-9-]+$/', $preset['id'])) {
                    $errors[] = "$path.id must contain only lowercase letters, numbers, and hyphens";
                }
            }

            // Check for duplicate IDs
            if (isset($preset['id']) && $preset['id'] !== '') {
                if (in_array($preset['id'], $presetIds, true)) {
                    $errors[] = "$path.id '{$preset['id']}' is already used";
                }
                $presetIds[] = $preset['id'];
            }

            // Validate format
            if (isset($preset['format'])) {
                if (!in_array($preset['format'], ['webp', 'jpg', 'png'], true)) {
                    $errors[] = "$path.format must be 'webp', 'jpg', or 'png'";
                }
            }

            // Check variants exist
            if (!isset($preset['variants']) || !is_array($preset['variants']) || count($preset['variants']) === 0) {
                $errors[] = "$path must have at least one variant";
            } else {
                $variantErrors = self::validateVariants($preset['variants'], "$path.variants");
                $errors = array_merge($errors, $variantErrors);
            }

            $index++;
        }

        return $errors;
    }

    private static function validateVariants(array $variants, string $pathPrefix): array
    {
        $errors = [];
        $variantIds = [];
        $index = 0;

        foreach ($variants as $variant) {
            $path = "$pathPrefix[$index]";

            // Check structure
            if (!isset($variant['id']) || !is_string($variant['id']) || $variant['id'] === '') {
                $errors[] = "$path.id is required and must be a non-empty string";
            }

            if (!isset($variant['label']) || !is_string($variant['label']) || $variant['label'] === '') {
                $errors[] = "$path.label is required and must be a non-empty string";
            }

            if (!isset($variant['width']) || !is_int($variant['width']) || $variant['width'] <= 0) {
                $errors[] = "$path.width must be a positive integer";
            }

            if (!isset($variant['height']) || !is_int($variant['height']) || $variant['height'] <= 0) {
                $errors[] = "$path.height must be a positive integer";
            }

            // Validate ID format
            if (isset($variant['id']) && $variant['id'] !== '') {
                if (!preg_match('/^[a-z0-9-]+$/', $variant['id'])) {
                    $errors[] = "$path.id must contain only lowercase letters, numbers, and hyphens";
                }
            }

            // Check for duplicate IDs
            if (isset($variant['id']) && $variant['id'] !== '') {
                if (in_array($variant['id'], $variantIds, true)) {
                    $errors[] = "$path.id '{$variant['id']}' is already used in this preset";
                }
                $variantIds[] = $variant['id'];
            }

            // Validate format if present
            if (isset($variant['format'])) {
                if (!in_array($variant['format'], ['webp', 'jpg', 'png'], true)) {
                    $errors[] = "$path.format must be 'webp', 'jpg', or 'png'";
                }
            }

            $index++;
        }

        return $errors;
    }

    public static function generateId(string $label): string
    {
        $id = strtolower($label);
        $id = preg_replace('/[^a-z0-9-]/', '-', $id);
        $id = preg_replace('/-+/', '-', $id);
        $id = trim($id, '-');
        return $id;
    }
}

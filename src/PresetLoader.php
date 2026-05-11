<?php

namespace App;

use RuntimeException;

class PresetLoader
{
    private static ?array $presets = null;

    public static function all(): array
    {
        if (self::$presets !== null) {
            return self::$presets;
        }

        $path = dirname(__DIR__) . '/config/presets.json';
        $json = file_get_contents($path);
        $data = json_decode($json, true);

        if (!isset($data['presets']) || !is_array($data['presets'])) {
            throw new RuntimeException('Invalid presets.json: missing "presets" array');
        }

        self::$presets = $data['presets'];
        return self::$presets;
    }

    public static function clearCache(): void
    {
        self::$presets = null;
    }

    public static function findPreset(string $presetId): ?array
    {
        foreach (self::all() as $preset) {
            if ($preset['id'] === $presetId) {
                return $preset;
            }
        }
        return null;
    }

    public static function findVariant(string $presetId, string $variantId): ?array
    {
        $preset = self::findPreset($presetId);
        if ($preset === null) {
            return null;
        }
        foreach ($preset['variants'] as $variant) {
            if ($variant['id'] === $variantId) {
                return array_merge($variant, ['format' => $variant['format'] ?? $preset['format'] ?? null]);
            }
        }
        return null;
    }
}

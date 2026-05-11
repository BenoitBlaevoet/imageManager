<?php

namespace App;

use RuntimeException;

class ComponentLoader
{
    private static ?array $components = null;

    public static function all(): array
    {
        if (self::$components !== null) {
            return self::$components;
        }

        $path = dirname(__DIR__) . '/config/components.json';
        $json = file_get_contents($path);
        $data = json_decode($json, true);

        if (!isset($data['components']) || !is_array($data['components'])) {
            throw new RuntimeException('Invalid components.json: missing "components" array');
        }

        self::$components = $data['components'];
        return self::$components;
    }

    public static function findComponent(string $id): ?array
    {
        foreach (self::all() as $component) {
            if ($component['id'] === $id) {
                return $component;
            }
        }
        return null;
    }

    public static function clearCache(): void
    {
        self::$components = null;
    }
}

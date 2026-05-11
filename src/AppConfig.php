<?php

namespace App;

use RuntimeException;

class AppConfig
{
    private static string $path = '';

    private static function path(): string
    {
        if (self::$path === '') {
            self::$path = dirname(__DIR__) . '/config/app-config.json';
        }
        return self::$path;
    }

    public static function get(): array
    {
        $json = file_get_contents(self::path());
        return json_decode($json, true) ?? [];
    }

    public static function outputFolder(): string
    {
        return self::get()['output_folder'] ?? '';
    }

    public static function defaultFormat(): string
    {
        return self::get()['default_format'] ?? 'webp';
    }

    public static function update(array $data): void
    {
        $current = self::get();
        $merged  = array_merge($current, $data);
        file_put_contents(self::path(), json_encode($merged, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES));
    }
}

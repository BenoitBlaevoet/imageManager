<?php

$root = dirname(__DIR__);

foreach (['data', 'storage/source', 'storage/cache', 'output'] as $dir) {
    $path = $root . '/' . $dir;
    if (!is_dir($path)) {
        mkdir($path, 0755, true);
    }
}

$configPath  = $root . '/config/app-config.json';
$examplePath = $root . '/config/app-config.example.json';

if (!file_exists($configPath)) {
    if (file_exists($examplePath)) {
        copy($examplePath, $configPath);
    } else {
        file_put_contents($configPath, json_encode([
            'output_folder' => '',
            'default_format' => 'webp',
        ], JSON_PRETTY_PRINT));
    }
}

$config = json_decode(file_get_contents($configPath), true) ?? [];

if (empty($config['output_folder'])) {
    $config['output_folder'] = $root . '/output';
    file_put_contents($configPath, json_encode($config, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES));
}

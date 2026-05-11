<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\PresetValidator;
use App\PresetLoader;

// Auth::requireAuth();

header('Content-Type: application/json');

$method = $_SERVER['REQUEST_METHOD'];

if ($method === 'GET') {
    try {
        $presets = PresetLoader::all();
        echo json_encode(['presets' => $presets]);
    } catch (\Exception $e) {
        http_response_code(500);
        echo json_encode(['error' => 'Failed to load presets']);
    }
    exit;
}

if ($method === 'POST') {
    $body = json_decode(file_get_contents('php://input'), true);

    if (!isset($body['presets']) || !is_array($body['presets'])) {
        http_response_code(400);
        echo json_encode(['error' => 'Missing or invalid presets array']);
        exit;
    }

    $errors = PresetValidator::validatePresets($body['presets']);

    if (!empty($errors)) {
        http_response_code(422);
        echo json_encode(['error' => 'Validation failed', 'errors' => $errors]);
        exit;
    }

    try {
        $path = dirname(__DIR__, 2) . '/config/presets.json';
        $json = json_encode(['presets' => $body['presets']], JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
        file_put_contents($path, $json);

        // Clear the static cache so next read is fresh
        PresetLoader::clearCache();

        echo json_encode(['success' => true]);
    } catch (\Exception $e) {
        http_response_code(500);
        echo json_encode(['error' => 'Failed to save presets']);
    }
    exit;
}

http_response_code(405);
echo json_encode(['error' => 'Method not allowed']);

<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\Database;
use App\ComponentLoader;
use App\ComponentValidator;
use App\ComponentRepository;
use App\AppConfig;

// Auth::requireAuth();

header('Content-Type: application/json');

$method = $_SERVER['REQUEST_METHOD'];

if ($method === 'GET') {
    try {
        $components = ComponentLoader::all();
        $db         = Database::get();
        $repo       = new ComponentRepository($db);
        $output     = [];

        foreach ($components as $component) {
            $activeImage  = $repo->activeImage($component['id']);
            $imageCount   = count($repo->forComponent($component['id']));
            $status       = $repo->generationStatus($component['id'], $component['presets'] ?? []);
            $outputFolder = AppConfig::outputFolder();

            $entry = [
                'id'                 => $component['id'],
                'label'              => $component['label'],
                'description'        => $component['description'] ?? '',
                'presets'            => $component['presets'] ?? [],
                'image_count'        => $imageCount,
                'variants_total'     => $status['total'],
                'variants_generated' => $status['generated'],
                'active_image'       => null,
            ];

            if ($activeImage) {
                $entry['active_image'] = [
                    'assignment_id' => (int) $activeImage['assignment_id'],
                    'id'            => $activeImage['id'],
                    'original_name' => $activeImage['original_name'],
                    'name'          => $activeImage['name'] ?: $activeImage['original_name'],
                    'filename'      => $activeImage['filename'],
                    'width'         => (int) $activeImage['width'],
                    'height'        => (int) $activeImage['height'],
                    'thumb_url'     => 'image-serve.php?file=' . urlencode($activeImage['filename']) . '&w=400&h=250&fit=crop',
                ];
            }

            $output[] = $entry;
        }

        echo json_encode(['components' => $output]);
    } catch (\Exception $e) {
        http_response_code(500);
        echo json_encode(['error' => $e->getMessage()]);
    }
    exit;
}

if ($method === 'POST') {
    $body = json_decode(file_get_contents('php://input'), true);

    if (!isset($body['components']) || !is_array($body['components'])) {
        http_response_code(400);
        echo json_encode(['error' => 'Missing or invalid components array']);
        exit;
    }

    $errors = ComponentValidator::validateComponents($body['components']);

    if (!empty($errors)) {
        http_response_code(422);
        echo json_encode(['error' => 'Validation failed', 'errors' => $errors]);
        exit;
    }

    $path = dirname(__DIR__, 2) . '/config/components.json';
    file_put_contents($path, json_encode(['components' => $body['components']], JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES));

    ComponentLoader::clearCache();

    echo json_encode(['success' => true]);
    exit;
}

http_response_code(405);
echo json_encode(['error' => 'Method not allowed']);

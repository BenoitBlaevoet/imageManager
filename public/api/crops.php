<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\Database;
use App\ImageRepository;
use App\CropRepository;
use App\PresetLoader;
use App\AppConfig;
use App\ComponentLoader;

// Auth::requireAuth();

header('Content-Type: application/json');

$method   = $_SERVER['REQUEST_METHOD'];
$db       = Database::get();
$images   = new ImageRepository($db);
$crops    = new CropRepository($db);

if ($method === 'GET') {
    $imageId     = $_GET['image_id']     ?? '';
    $componentId = $_GET['component_id'] ?? '';

    if ($imageId === '') {
        http_response_code(400);
        echo json_encode(['error' => 'Missing image_id']);
        exit;
    }

    $image = $images->findById($imageId);
    if ($image === null) {
        http_response_code(404);
        echo json_encode(['error' => 'Image not found']);
        exit;
    }

    $savedCrops = [];
    foreach ($crops->forImage($imageId) as $row) {
        $savedCrops[$row['preset_id'] . ':' . $row['variant_id']] = $row;
    }

    $outputFolder  = AppConfig::outputFolder();
    $defaultFormat = AppConfig::defaultFormat();
    $result        = [];

    // If a component_id is given, only show that component's presets
    $allPresets = PresetLoader::all();
    if ($componentId !== '') {
        $component = ComponentLoader::findComponent($componentId);
        if ($component !== null) {
            $allowedPresetIds = $component['presets'] ?? [];
            $allPresets = array_values(array_filter($allPresets, fn($p) => in_array($p['id'], $allowedPresetIds, true)));
        }
    }

    foreach ($allPresets as $preset) {
        $presetFormat  = $preset['format'] ?? $defaultFormat;
        $variantStates = [];

        foreach ($preset['variants'] as $variant) {
            $key    = $preset['id'] . ':' . $variant['id'];
            $saved  = $savedCrops[$key] ?? null;
            $format = $variant['format'] ?? $presetFormat;
            $ext    = $format;
            $base   = $componentId !== ''
                ? $componentId . '_' . $variant['id']
                : $imageId . '_' . $preset['id'] . '_' . $variant['id'];

            $file1x = $base . '_1x.' . $ext;
            $file2x = $base . '_2x.' . $ext;

            $variantStates[] = [
                'id'           => $variant['id'],
                'label'        => $variant['label'],
                'width'        => $variant['width'],
                'height'       => $variant['height'],
                'format'       => $format,
                'crop'         => $saved ? [
                    'x' => (int) $saved['crop_x'],
                    'y' => (int) $saved['crop_y'],
                    'w' => (int) $saved['crop_w'],
                    'h' => (int) $saved['crop_h'],
                ] : null,
                'generated_at' => $saved['generated_at'] ?? null,
                'files'        => [
                    '1x' => $file1x,
                    '2x' => $file2x,
                ],
                'files_exist'  => $outputFolder !== '' && file_exists($outputFolder . '/' . $file1x),
            ];
        }

        $result[] = [
            'id'       => $preset['id'],
            'label'    => $preset['label'],
            'variants' => $variantStates,
        ];
    }

    echo json_encode(['presets' => $result]);
    exit;
}

if ($method === 'POST') {
    $body = json_decode(file_get_contents('php://input'), true);

    $imageId   = $body['image_id']   ?? '';
    $presetId  = $body['preset_id']  ?? '';
    $variantId = $body['variant_id'] ?? '';
    $x         = (int) ($body['crop_x'] ?? 0);
    $y         = (int) ($body['crop_y'] ?? 0);
    $w         = (int) ($body['crop_w'] ?? 0);
    $h         = (int) ($body['crop_h'] ?? 0);

    if ($imageId === '' || $presetId === '' || $variantId === '') {
        http_response_code(400);
        echo json_encode(['error' => 'Missing required fields']);
        exit;
    }

    if ($images->findById($imageId) === null) {
        http_response_code(404);
        echo json_encode(['error' => 'Image not found']);
        exit;
    }

    if (PresetLoader::findVariant($presetId, $variantId) === null) {
        http_response_code(400);
        echo json_encode(['error' => 'Unknown preset/variant']);
        exit;
    }

    if ($w <= 0 || $h <= 0) {
        http_response_code(422);
        echo json_encode(['error' => 'Invalid crop dimensions']);
        exit;
    }

    $crops->upsert($imageId, $presetId, $variantId, $x, $y, $w, $h);

    echo json_encode(['success' => true]);
    exit;
}

http_response_code(405);
echo json_encode(['error' => 'Method not allowed']);

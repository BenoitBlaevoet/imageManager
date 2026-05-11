<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\Database;
use App\ImageRepository;
use App\CropRepository;
use App\PresetLoader;
use App\GlideServer;
use App\AppConfig;
use App\ComponentLoader;

header('Content-Type: application/json');

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    echo json_encode(['error' => 'Method not allowed']);
    exit;
}

$body        = json_decode(file_get_contents('php://input'), true);
$imageId     = $body['image_id']    ?? '';
$presetId    = $body['preset_id']   ?? '';
$variantId   = $body['variant_id']  ?? null;
$componentId = $body['component_id'] ?? null;

if ($imageId === '' || $presetId === '') {
    http_response_code(400);
    echo json_encode(['error' => 'Missing image_id or preset_id']);
    exit;
}

$db     = Database::get();
$images = new ImageRepository($db);
$crops  = new CropRepository($db);

$image = $images->findById($imageId);
if ($image === null) {
    http_response_code(404);
    echo json_encode(['error' => 'Image not found']);
    exit;
}

$preset = PresetLoader::findPreset($presetId);
if ($preset === null) {
    http_response_code(400);
    echo json_encode(['error' => 'Unknown preset']);
    exit;
}

$outputFolder  = AppConfig::outputFolder();
$defaultFormat = AppConfig::defaultFormat();

if ($outputFolder === '' || !is_dir($outputFolder)) {
    http_response_code(400);
    echo json_encode(['error' => 'Output folder not configured or does not exist']);
    exit;
}

// Build output directory: {outputFolder}/{component_id}/{image_slug}/
// Falls back to {outputFolder}/{image_slug}/ when no component context
$imageSlug = slugify($image['name'] ?: $image['original_name']);
$subDir    = $componentId !== null
    ? $outputFolder . '/' . $componentId . '/' . $imageSlug
    : $outputFolder . '/' . $imageSlug;

if (!is_dir($subDir)) {
    mkdir($subDir, 0755, true);
}

$server    = GlideServer::create();
$generated = [];

$variants = $variantId !== null
    ? array_filter($preset['variants'], fn($v) => $v['id'] === $variantId)
    : $preset['variants'];

foreach ($variants as $variant) {
    $savedCrop = $crops->find($imageId, $presetId, $variant['id']);

    if ($savedCrop !== null) {
        $cropX = (int) $savedCrop['crop_x'];
        $cropY = (int) $savedCrop['crop_y'];
        $cropW = (int) $savedCrop['crop_w'];
        $cropH = (int) $savedCrop['crop_h'];
    } else {
        $ratio = $variant['width'] / $variant['height'];
        $srcW  = $image['width'];
        $srcH  = $image['height'];

        if ($srcW / $srcH > $ratio) {
            $cropH = $srcH;
            $cropW = (int) round($srcH * $ratio);
        } else {
            $cropW = $srcW;
            $cropH = (int) round($srcW / $ratio);
        }

        $cropX = (int) round(($srcW - $cropW) / 2);
        $cropY = (int) round(($srcH - $cropH) / 2);
    }

    $format  = $variant['format'] ?? $preset['format'] ?? $defaultFormat;
    $cropStr = "{$cropW},{$cropH},{$cropX},{$cropY}";
    $base    = $presetId . '_' . $variant['id'];
    $files   = [];

    foreach ([1, 2] as $dpr) {
        $w       = $variant['width']  * $dpr;
        $h       = $variant['height'] * $dpr;
        $suffix  = $dpr === 1 ? '1x' : '2x';
        $outName = $base . '_' . $suffix . '.' . $format;
        $outPath = $subDir . '/' . $outName;

        $params = [
            'w'    => $w,
            'h'    => $h,
            'fit'  => 'crop',
            'crop' => $cropStr,
            'fm'   => $format,
        ];

        $cachePath = $server->makeImage($image['filename'], $params);
        $contents  = $server->getCache()->read($cachePath);
        file_put_contents($outPath, $contents);

        $files[$suffix] = $outName;
    }

    $crops->upsert($imageId, $presetId, $variant['id'], $cropX, $cropY, $cropW, $cropH);
    $crops->markGenerated($imageId, $presetId, $variant['id']);

    $generated[] = [
        'variant_id' => $variant['id'],
        'files'      => $files,
        'folder'     => $subDir,
    ];
}

echo json_encode(['generated' => $generated]);

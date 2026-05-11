<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\AppConfig;
use App\Database;
use App\ImageRepository;

header('Content-Type: application/json');

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    echo json_encode(['error' => 'Method not allowed']);
    exit;
}

$outputFolder = AppConfig::outputFolder();

if ($outputFolder === '') {
    http_response_code(400);
    echo json_encode(['error' => 'Output folder not configured']);
    exit;
}

if (!is_dir($outputFolder)) {
    http_response_code(400);
    echo json_encode(['error' => 'Output folder does not exist']);
    exit;
}

// Optionally open a component/image subfolder
$body        = json_decode(file_get_contents('php://input'), true) ?? [];
$componentId = $body['component_id'] ?? '';
$imageId     = $body['image_id']     ?? '';
$targetFolder = $outputFolder;

if ($componentId !== '' && $imageId !== '') {
    $image = (new ImageRepository(Database::get()))->findById($imageId);
    if ($image !== null) {
        $slug = slugify($image['name'] ?: $image['original_name']);
        $sub  = $outputFolder . '/' . $componentId . '/' . $slug;
        if (is_dir($sub)) {
            $targetFolder = $sub;
        }
    }
}

match (PHP_OS_FAMILY) {
    'Windows' => exec('explorer ' . escapeshellarg(str_replace('/', '\\', $targetFolder))),
    'Darwin'  => exec('open ' . escapeshellarg($targetFolder)),
    default   => exec('xdg-open ' . escapeshellarg($targetFolder)),
};

echo json_encode(['success' => true]);

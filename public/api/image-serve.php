<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\GlideServer;
use App\Database;
use App\ImageRepository;

// Auth::requireAuth();

$file = $_GET['file'] ?? '';

if (!preg_match('/^img_[a-zA-Z0-9._-]+$/', $file)) {
    http_response_code(400);
    echo 'Invalid filename';
    exit;
}

$repo  = new ImageRepository(Database::get());
$image = $repo->findByFilename($file);
if ($image === null) {
    http_response_code(404);
    echo 'Image not found';
    exit;
}

try {
    $server = GlideServer::create();
    $server->outputImage($file, $_GET);
} catch (\Exception $e) {
    http_response_code(404);
    echo 'Image not found';
}

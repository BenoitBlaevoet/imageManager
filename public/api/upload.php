<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\Database;
use App\ImageRepository;
use App\ComponentRepository;
use App\ComponentLoader;
use App\TagRepository;

header('Content-Type: application/json');

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    echo json_encode(['error' => 'Method not allowed']);
    exit;
}

if (empty($_FILES['image'])) {
    http_response_code(400);
    echo json_encode(['error' => 'No file uploaded']);
    exit;
}

$file    = $_FILES['image'];
$maxSize = 20 * 1024 * 1024; // 20 MB

if ($file['size'] > $maxSize) {
    http_response_code(413);
    echo json_encode(['error' => 'File too large (max 20 MB)']);
    exit;
}

$finfo    = new finfo(FILEINFO_MIME_TYPE);
$mimeType = $finfo->file($file['tmp_name']);
$allowed  = ['image/jpeg', 'image/png', 'image/webp'];

if (!in_array($mimeType, $allowed, true)) {
    http_response_code(415);
    echo json_encode(['error' => 'Unsupported file type']);
    exit;
}

$extMap   = ['image/jpeg' => 'jpg', 'image/png' => 'png', 'image/webp' => 'webp'];
$ext      = $extMap[$mimeType];
$id       = uniqid('img_', true);
$filename = $id . '.' . $ext;
$dest     = dirname(__DIR__, 2) . '/storage/source/' . $filename;

if (!move_uploaded_file($file['tmp_name'], $dest)) {
    http_response_code(500);
    echo json_encode(['error' => 'Failed to save file']);
    exit;
}

[$width, $height] = getimagesize($dest);

// Resolve display name
$rawName     = trim($_POST['name'] ?? '');
$displayName = $rawName !== '' ? $rawName : pathinfo($file['name'], PATHINFO_FILENAME);

$db   = Database::get();
$repo = new ImageRepository($db);
$repo->insert([
    'id'            => $id,
    'original_name' => $file['name'],
    'name'          => $displayName,
    'filename'      => $filename,
    'mime_type'     => $mimeType,
    'width'         => $width,
    'height'        => $height,
    'file_size'     => $file['size'],
]);

// Handle tags
$tagRepo = new TagRepository($db);
$tagsJson = $_POST['tags'] ?? '[]';
$tagIds   = json_decode($tagsJson, true);
if (is_array($tagIds) && !empty($tagIds)) {
    $tagRepo->setTagsForImage($id, array_map('intval', $tagIds));
}

// Auto-assign to component if provided
$componentId = $_POST['component_id'] ?? '';
if ($componentId !== '' && ComponentLoader::findComponent($componentId) !== null) {
    (new ComponentRepository($db))->assign($componentId, $id);
}

$tags = $tagRepo->tagsForImage($id);

echo json_encode([
    'id'            => $id,
    'filename'      => $filename,
    'original_name' => $file['name'],
    'name'          => $displayName,
    'width'         => $width,
    'height'        => $height,
    'tags'          => $tags,
    'component_id'  => $componentId ?: null,
]);

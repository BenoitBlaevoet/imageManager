<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\Database;
use App\ImageRepository;
use App\TagRepository;
use App\ComponentRepository;
use App\AppConfig;

header('Content-Type: application/json');

$db     = Database::get();
$repo   = new ImageRepository($db);
$method = $_SERVER['REQUEST_METHOD'];

if ($method === 'GET') {
    $search = trim($_GET['search'] ?? '');
    $tagIds = array_map('intval', $_GET['tags'] ?? []);
    $tagIds = array_filter($tagIds);

    $images  = $repo->findAll($search, $tagIds);
    $tagRepo = new TagRepository($db);
    $tagMap  = $tagRepo->tagsForImages(array_column($images, 'id'));

    foreach ($images as &$img) {
        $img['name']      = $img['name'] ?: $img['original_name'];
        $img['tags']      = $tagMap[$img['id']] ?? [];
        $img['thumb_url'] = 'image-serve.php?file=' . urlencode($img['filename']) . '&w=300&h=200&fit=crop';
    }
    unset($img);

    echo json_encode($images);
    exit;
}

if ($method === 'DELETE') {
    $id = $_GET['id'] ?? '';
    if ($id === '') {
        http_response_code(400);
        echo json_encode(['error' => 'Missing id']);
        exit;
    }

    $image = $repo->findById($id);
    if ($image === null) {
        http_response_code(404);
        echo json_encode(['error' => 'Image not found']);
        exit;
    }

    // Delete source file
    $sourcePath = dirname(__DIR__, 2) . '/storage/source/' . $image['filename'];
    if (file_exists($sourcePath)) {
        unlink($sourcePath);
    }

    // Delete structured output folders ({outputFolder}/{component_id}/{image_slug}/)
    $outputFolder = AppConfig::outputFolder();
    if ($outputFolder !== '' && is_dir($outputFolder)) {
        $imageSlug    = slugify($image['name'] ?: $image['original_name']);
        $assignments  = (new ComponentRepository($db))->allAssignmentsForImage($id);

        foreach ($assignments as $assignment) {
            $dir = $outputFolder . '/' . $assignment['component_id'] . '/' . $imageSlug;
            if (is_dir($dir)) {
                foreach (glob($dir . '/*') ?: [] as $file) {
                    unlink($file);
                }
                @rmdir($dir);
            }
        }

        // Also clean up any legacy flat files from old naming scheme
        foreach (glob($outputFolder . '/' . $id . '_*') ?: [] as $file) {
            if (is_file($file)) {
                unlink($file);
            }
        }
    }

    // Delete Glide cache for this image
    $cacheDir = dirname(__DIR__, 2) . '/storage/cache/' . $image['filename'];
    if (is_dir($cacheDir)) {
        array_map('unlink', glob($cacheDir . '/*') ?: []);
        rmdir($cacheDir);
    }

    $repo->delete($id);

    echo json_encode(['success' => true]);
    exit;
}

http_response_code(405);
echo json_encode(['error' => 'Method not allowed']);

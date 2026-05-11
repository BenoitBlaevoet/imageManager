<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\Database;
use App\ComponentRepository;
use App\ImageRepository;
use App\TagRepository;
use App\ComponentLoader;

header('Content-Type: application/json');

$method = $_SERVER['REQUEST_METHOD'];
$db     = Database::get();
$repo   = new ComponentRepository($db);

if ($method === 'GET') {
    $componentId = $_GET['component_id'] ?? '';
    if ($componentId === '') {
        http_response_code(400);
        echo json_encode(['error' => 'Missing component_id']);
        exit;
    }

    $rows    = $repo->forComponent($componentId);
    $tagRepo = new TagRepository($db);
    $imageIds = array_column($rows, 'id');
    $tagMap   = $tagRepo->tagsForImages($imageIds);

    $images = array_map(fn($r) => [
        'assignment_id' => (int) $r['assignment_id'],
        'is_active'     => (bool) $r['is_active'],
        'assigned_at'   => $r['assigned_at'],
        'id'            => $r['id'],
        'original_name' => $r['original_name'],
        'name'          => $r['name'] ?: $r['original_name'],
        'filename'      => $r['filename'],
        'width'         => (int) $r['width'],
        'height'        => (int) $r['height'],
        'file_size'     => (int) $r['file_size'],
        'tags'          => $tagMap[$r['id']] ?? [],
        'thumb_url'     => 'image-serve.php?file=' . urlencode($r['filename']) . '&w=300&h=200&fit=crop',
    ], $rows);

    echo json_encode(['images' => $images]);
    exit;
}

if ($method === 'POST') {
    $body        = json_decode(file_get_contents('php://input'), true);
    $componentId = $body['component_id'] ?? '';
    $imageId     = $body['image_id']     ?? '';

    if ($componentId === '' || $imageId === '') {
        http_response_code(400);
        echo json_encode(['error' => 'Missing component_id or image_id']);
        exit;
    }

    if (ComponentLoader::findComponent($componentId) === null) {
        http_response_code(404);
        echo json_encode(['error' => 'Component not found']);
        exit;
    }

    $imgRepo = new ImageRepository($db);
    if ($imgRepo->findById($imageId) === null) {
        http_response_code(404);
        echo json_encode(['error' => 'Image not found']);
        exit;
    }

    $repo->assign($componentId, $imageId);
    echo json_encode(['success' => true]);
    exit;
}

if ($method === 'PATCH') {
    $body         = json_decode(file_get_contents('php://input'), true);
    $assignmentId = (int) ($body['assignment_id'] ?? 0);

    if ($assignmentId === 0) {
        http_response_code(400);
        echo json_encode(['error' => 'Missing assignment_id']);
        exit;
    }

    $repo->setActive($assignmentId);
    echo json_encode(['success' => true]);
    exit;
}

if ($method === 'DELETE') {
    $assignmentId = (int) ($_GET['id'] ?? 0);

    if ($assignmentId === 0) {
        http_response_code(400);
        echo json_encode(['error' => 'Missing id']);
        exit;
    }

    $repo->removeAssignment($assignmentId);
    echo json_encode(['success' => true]);
    exit;
}

http_response_code(405);
echo json_encode(['error' => 'Method not allowed']);

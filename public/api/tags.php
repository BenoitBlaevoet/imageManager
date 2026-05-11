<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\Database;
use App\TagRepository;

header('Content-Type: application/json');

$method = $_SERVER['REQUEST_METHOD'];
$db     = Database::get();
$repo   = new TagRepository($db);

if ($method === 'GET') {
    echo json_encode(['tags' => $repo->all()]);
    exit;
}

if ($method === 'POST') {
    $body = json_decode(file_get_contents('php://input'), true);
    $name = trim($body['name'] ?? '');

    if ($name === '') {
        http_response_code(400);
        echo json_encode(['error' => 'Tag name is required']);
        exit;
    }

    $existing = $repo->findByName($name);
    if ($existing) {
        echo json_encode(['id' => (int) $existing['id'], 'name' => $existing['name']]);
        exit;
    }

    $id = $repo->create($name);
    echo json_encode(['id' => $id, 'name' => $name]);
    exit;
}

if ($method === 'DELETE') {
    $id = (int) ($_GET['id'] ?? 0);
    if ($id === 0) {
        http_response_code(400);
        echo json_encode(['error' => 'Missing id']);
        exit;
    }
    $repo->delete($id);
    echo json_encode(['success' => true]);
    exit;
}

http_response_code(405);
echo json_encode(['error' => 'Method not allowed']);

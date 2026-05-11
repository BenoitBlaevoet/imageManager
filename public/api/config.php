<?php

require dirname(__DIR__, 2) . '/vendor/autoload.php';

use App\AppConfig;

// Auth::requireAuth();

header('Content-Type: application/json');

$method = $_SERVER['REQUEST_METHOD'];

if ($method === 'GET') {
    echo json_encode(AppConfig::get());
    exit;
}

if ($method === 'POST') {
    $body = json_decode(file_get_contents('php://input'), true);

    $updates = [];

    if (isset($body['output_folder'])) {
        $folder = rtrim(str_replace('\\', '/', $body['output_folder']), '/');

        if ($folder === '') {
            $folder = str_replace('\\', '/', dirname(__DIR__, 2) . '/output');
        }

        if (!is_dir($folder)) {
            http_response_code(422);
            echo json_encode(['error' => 'Output folder does not exist: ' . $folder]);
            exit;
        }

        $updates['output_folder'] = $folder;
    }

    if (isset($body['default_format'])) {
        $fmt = $body['default_format'];
        if (!in_array($fmt, ['webp', 'jpg', 'png'], true)) {
            http_response_code(422);
            echo json_encode(['error' => 'Invalid format. Allowed: webp, jpg, png']);
            exit;
        }
        $updates['default_format'] = $fmt;
    }

    if (!empty($updates)) {
        AppConfig::update($updates);
    }

    echo json_encode(AppConfig::get());
    exit;
}

http_response_code(405);
echo json_encode(['error' => 'Method not allowed']);

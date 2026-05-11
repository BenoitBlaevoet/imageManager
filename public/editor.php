<?php
require dirname(__DIR__) . '/vendor/autoload.php';

use App\Database;
use App\ImageRepository;
use App\ComponentLoader;

// Auth::requireUser();

$imageId     = $_GET['image_id']    ?? ($_GET['id'] ?? '');
$componentId = $_GET['component_id'] ?? '';

$repo  = new ImageRepository(Database::get());
$image = $imageId !== '' ? $repo->findById($imageId) : null;

if ($image === null) {
    header('Location: /');
    exit;
}

$component = $componentId !== '' ? ComponentLoader::findComponent($componentId) : null;

$originalWidth  = (int) $image['width'];
$originalHeight = (int) $image['height'];
$filename       = htmlspecialchars($image['filename']);
$originalName   = htmlspecialchars($image['original_name']);
$compLabel      = $component ? htmlspecialchars($component['label']) : '';
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Edit<?= $compLabel ? ' — ' . $compLabel : '' ?> — <?= $originalName ?></title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/cropperjs/1.6.2/cropper.min.css">
    <link rel="stylesheet" href="assets/css/app.css">
    <link rel="stylesheet" href="assets/css/editor.css">
</head>
<body class="editor-layout">

<header class="topbar">
    <a href="/" class="topbar-back">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
        Library
    </a>
    <div class="topbar-brand">
        <?php if ($compLabel): ?>
            <span class="topbar-component-tag"><?= $compLabel ?></span>
        <?php endif ?>
        <?= $originalName ?>
    </div>
    <div class="topbar-meta"><?= $originalWidth ?> × <?= $originalHeight ?> px</div>
</header>

<div class="editor-body"
     data-image-id="<?= htmlspecialchars($imageId) ?>"
     data-filename="<?= $filename ?>"
     data-original-width="<?= $originalWidth ?>"
     data-original-height="<?= $originalHeight ?>"
     data-component-id="<?= htmlspecialchars($componentId) ?>">

    <!-- Left panel: preset navigator -->
    <aside class="preset-panel" id="preset-panel">
        <div id="preset-list">
            <div class="loading-state">Loading presets…</div>
        </div>
    </aside>

    <!-- Right panel: crop canvas -->
    <section class="crop-panel" id="crop-panel">
        <div class="crop-panel-empty" id="crop-panel-empty">
            <p>Select a variant on the left to start cropping.</p>
        </div>
        <div class="crop-panel-active" id="crop-panel-active" style="display:none">
            <div class="crop-toolbar">
                <span class="crop-toolbar-info" id="crop-info-label"></span>
                <button class="btn btn-primary" id="btn-save-crop">Save Crop</button>
            </div>
            <div class="crop-canvas-wrap" id="crop-canvas-wrap">
                <img id="crop-img" src="" alt="">
            </div>
            <div class="crop-coords" id="crop-coords">
                <span>x: <strong id="coord-x">—</strong></span>
                <span>y: <strong id="coord-y">—</strong></span>
                <span>w: <strong id="coord-w">—</strong></span>
                <span>h: <strong id="coord-h">—</strong></span>
            </div>
        </div>
    </section>

</div>

<script src="https://cdnjs.cloudflare.com/ajax/libs/cropperjs/1.6.2/cropper.min.js"></script>
<script src="assets/js/editor.js"></script>
</body>
</html>

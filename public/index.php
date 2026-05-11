<?php
require dirname(__DIR__) . '/vendor/autoload.php';
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Image Manager</title>
    <link rel="stylesheet" href="assets/css/app.css">
</head>
<body>

<header class="topbar">
    <div class="topbar-brand">Image Manager</div>
    <nav class="topbar-tabs">
        <button class="topbar-tab active" id="tab-components">Components</button>
        <button class="topbar-tab" id="tab-images">Images</button>
    </nav>
    <div class="topbar-actions">
        <button class="btn btn-secondary" id="btn-components">Manage Components</button>
        <button class="btn btn-secondary" id="btn-presets">Manage Presets</button>
        <button class="btn btn-secondary" id="btn-settings">Settings</button>
    </div>
</header>

<!-- Components section -->
<section id="section-components" class="library">
    <div class="component-grid" id="component-grid">
        <div class="loading-state">Loading components…</div>
    </div>
</section>

<!-- Images section -->
<section id="section-images" class="library" style="display:none">
    <div class="library-toolbar">
        <input type="search" id="images-search" class="field-input library-search"
               placeholder="Search by name or filename…">
        <button class="btn btn-primary" id="btn-upload-image">Upload Image</button>
        <button class="btn btn-secondary" id="btn-manage-tags">Tags</button>
    </div>
    <div id="images-tag-filter" class="tag-filter-bar"></div>
    <div class="image-library-grid" id="image-library-grid">
        <div class="loading-state">Loading images…</div>
    </div>
</section>

<!-- ─── Modals ─────────────────────────────────────────────────────────────── -->

<!-- Upload modal (library + component context) -->
<div class="modal-overlay" id="upload-modal" style="display:none">
    <div class="modal">
        <div class="modal-header">
            <h2>Upload Image</h2>
            <button class="modal-close" id="upload-close">&times;</button>
        </div>
        <div class="modal-body">
            <p id="upload-context-label" class="field-hint" style="display:none;margin-bottom:.5rem"></p>

            <!-- Step 1: Drop zone -->
            <div id="upload-step-1">
                <div class="drop-zone" id="drop-zone">
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                    <p>Drag &amp; drop an image here</p>
                    <p class="hint">or</p>
                    <label class="btn btn-secondary" for="file-input">Browse file</label>
                    <input type="file" id="file-input" accept="image/jpeg,image/png,image/webp" style="display:none">
                </div>
            </div>

            <!-- Step 2: Metadata form -->
            <div id="upload-step-2" style="display:none">
                <div class="form-row">
                    <label class="field-label" for="upload-name">Name</label>
                    <input type="text" id="upload-name" class="field-input" placeholder="Image display name">
                </div>
                <div class="form-row">
                    <label class="field-label">Tags</label>
                    <div id="upload-tags-list" class="tag-checkbox-list"></div>
                    <div class="tag-new-row">
                        <input type="text" id="upload-new-tag" class="field-input" placeholder="New tag name…">
                        <button type="button" class="btn btn-secondary btn-sm" id="btn-add-tag-upload">Add</button>
                    </div>
                </div>
            </div>

            <!-- Progress -->
            <div class="upload-progress" id="upload-progress" style="display:none">
                <div class="progress-bar"><div class="progress-fill" id="progress-fill"></div></div>
                <p id="progress-label">Uploading…</p>
            </div>
        </div>
        <div class="modal-footer" id="upload-footer" style="display:none">
            <button class="btn btn-ghost" id="btn-upload-back">← Back</button>
            <button class="btn btn-primary" id="btn-do-upload">Upload</button>
        </div>
    </div>
</div>

<!-- Image picker modal (add image to component) -->
<div class="modal-overlay" id="image-picker-modal" style="display:none">
    <div class="modal modal-wide" style="max-height:85vh;display:flex;flex-direction:column">
        <div class="modal-header">
            <h2 id="image-picker-title">Add Image</h2>
            <button class="modal-close" id="image-picker-close">&times;</button>
        </div>
        <div class="modal-body" style="display:flex;flex-direction:column;gap:.75rem;overflow:hidden;flex:1">
            <div class="picker-toolbar">
                <input type="search" id="picker-search" class="field-input picker-search"
                       placeholder="Search images…">
                <div id="picker-tag-filters" class="tag-filter-bar" style="margin-top:.5rem"></div>
            </div>
            <div class="image-browser-grid" id="picker-image-grid"
                 style="overflow-y:auto;flex:1;min-height:0">
                <div class="loading-state">Loading…</div>
            </div>
        </div>
        <div class="modal-footer" style="justify-content:space-between">
            <button class="btn btn-secondary" id="btn-picker-upload">Upload New</button>
            <button class="btn btn-ghost" id="btn-picker-cancel">Cancel</button>
        </div>
    </div>
</div>

<!-- Tag manager modal -->
<div class="modal-overlay" id="tag-manager-modal" style="display:none">
    <div class="modal">
        <div class="modal-header">
            <h2>Manage Tags</h2>
            <button class="modal-close" id="tag-manager-close">&times;</button>
        </div>
        <div class="modal-body">
            <div id="tag-manager-list" style="max-height:300px;overflow-y:auto;margin-bottom:.75rem"></div>
            <div class="tag-new-row">
                <input type="text" id="new-tag-input" class="field-input" placeholder="New tag name…">
                <button type="button" class="btn btn-primary btn-sm" id="btn-create-tag">Create</button>
            </div>
        </div>
    </div>
</div>

<!-- Image browser modal (component assigned images) -->
<div class="modal-overlay" id="image-browser-modal" style="display:none">
    <div class="modal modal-wide">
        <div class="modal-header">
            <h2 id="image-browser-title">Images</h2>
            <button class="modal-close" id="image-browser-close">&times;</button>
        </div>
        <div class="modal-body" style="display:block;overflow-y:auto;max-height:60vh">
            <div class="image-browser-grid" id="image-browser-grid">
                <div class="loading-state">Loading…</div>
            </div>
        </div>
    </div>
</div>

<!-- Settings modal -->
<div class="modal-overlay" id="settings-modal" style="display:none">
    <div class="modal">
        <div class="modal-header">
            <h2>Settings</h2>
            <button class="modal-close" id="settings-close">&times;</button>
        </div>
        <div class="modal-body">
            <label class="field-label" for="output-folder">Output folder (absolute path)</label>
            <input type="text" id="output-folder" class="field-input"
                   placeholder="e.g. C:/projects/my-site/public/images">
            <p class="field-hint">Generated images saved here as <code>{component}/{image-name}/{preset}_{variant}_1x.webp</code>.</p>

            <label class="field-label" style="margin-top:.75rem" for="default-format">Default output format</label>
            <select id="default-format" class="field-input">
                <option value="webp">WebP</option>
                <option value="jpg">JPEG</option>
                <option value="png">PNG</option>
            </select>

            <div class="modal-footer">
                <button class="btn btn-primary" id="btn-save-settings">Save</button>
            </div>
        </div>
    </div>
</div>

<!-- Presets Manager modal -->
<div class="modal-overlay" id="presets-modal" style="display:none">
    <div class="modal modal-wide">
        <div class="modal-header">
            <h2>Manage Presets</h2>
            <button class="modal-close" id="presets-close">&times;</button>
        </div>
        <div class="modal-body modal-body-two-col">
            <div class="presets-panel-left">
                <h3>Presets</h3>
                <div class="preset-tree" id="preset-tree">
                    <div class="loading-state">Loading presets…</div>
                </div>
                <button class="btn btn-secondary" id="btn-add-preset" style="width:100%;margin-top:0.75rem">Add Preset</button>
            </div>
            <div class="presets-panel-right">
                <div id="preset-form-empty" class="preset-form-empty">
                    <p>Select a preset to edit, or add a new one.</p>
                </div>
                <form id="preset-edit-form" class="preset-edit-form" style="display:none">
                    <h3 id="form-title">Edit Preset</h3>
                    <div id="preset-fields">
                        <div class="form-row">
                            <label class="field-label" for="form-preset-id">Preset ID</label>
                            <input type="text" id="form-preset-id" class="field-input" placeholder="e.g. homepage-banner" required>
                        </div>
                        <div class="form-row">
                            <label class="field-label" for="form-preset-label">Preset Name</label>
                            <input type="text" id="form-preset-label" class="field-input" placeholder="e.g. Homepage Banner" required>
                        </div>
                        <div class="form-row">
                            <label class="field-label" for="form-preset-format">Format (optional)</label>
                            <select id="form-preset-format" class="field-input">
                                <option value="">Inherit from variant</option>
                                <option value="webp">WebP</option>
                                <option value="jpg">JPEG</option>
                                <option value="png">PNG</option>
                            </select>
                        </div>
                    </div>
                    <div id="variants-section">
                        <h4>Variants</h4>
                        <div id="variants-list"></div>
                        <button type="button" class="btn btn-secondary" id="btn-add-variant">Add Variant</button>
                    </div>
                    <div style="border-top:1px solid var(--border);margin-top:1rem;padding-top:1rem">
                        <button type="button" class="btn btn-danger" id="btn-delete-preset">Delete Preset</button>
                    </div>
                </form>
            </div>
        </div>
        <div class="modal-footer">
            <button type="button" class="btn btn-ghost" id="btn-close-presets">Cancel</button>
            <button type="button" class="btn btn-primary" id="btn-save-presets">Save All Changes</button>
        </div>
    </div>
</div>

<!-- Components Manager modal -->
<div class="modal-overlay" id="components-modal" style="display:none">
    <div class="modal modal-wide">
        <div class="modal-header">
            <h2>Manage Components</h2>
            <button class="modal-close" id="components-close">&times;</button>
        </div>
        <div class="modal-body modal-body-two-col">
            <div class="presets-panel-left">
                <h3>Components</h3>
                <div class="preset-tree" id="component-tree">
                    <div class="loading-state">Loading…</div>
                </div>
                <button class="btn btn-secondary" id="btn-add-component" style="width:100%;margin-top:0.75rem">Add Component</button>
            </div>
            <div class="presets-panel-right">
                <div id="component-form-empty" class="preset-form-empty">
                    <p>Select a component to edit, or add a new one.</p>
                </div>
                <form id="component-edit-form" class="preset-edit-form" style="display:none">
                    <h3 id="component-form-title">Edit Component</h3>
                    <div class="form-row">
                        <label class="field-label" for="form-component-id">Component ID</label>
                        <input type="text" id="form-component-id" class="field-input" placeholder="e.g. partner-banner" required>
                    </div>
                    <div class="form-row">
                        <label class="field-label" for="form-component-label">Name</label>
                        <input type="text" id="form-component-label" class="field-input" placeholder="e.g. Partner Page Banner" required>
                    </div>
                    <div class="form-row">
                        <label class="field-label" for="form-component-description">Description (optional)</label>
                        <input type="text" id="form-component-description" class="field-input" placeholder="Where this image is used">
                    </div>
                    <div class="form-row" id="component-presets-section">
                        <label class="field-label">Presets (select at least one)</label>
                        <div id="component-preset-checkboxes" class="preset-checkbox-list"></div>
                    </div>
                    <div style="border-top:1px solid var(--border);margin-top:1rem;padding-top:1rem">
                        <button type="button" class="btn btn-danger" id="btn-delete-component">Delete Component</button>
                    </div>
                </form>
            </div>
        </div>
        <div class="modal-footer">
            <button type="button" class="btn btn-ghost" id="btn-close-components">Cancel</button>
            <button type="button" class="btn btn-primary" id="btn-save-components">Save All Changes</button>
        </div>
    </div>
</div>

<script src="assets/js/images.js"></script>
<script src="assets/js/library.js"></script>
</body>
</html>

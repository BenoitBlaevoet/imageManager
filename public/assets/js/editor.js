(() => {
    const editorBody   = document.querySelector('.editor-body');
    const IMAGE_ID     = editorBody.dataset.imageId;
    const FILENAME     = editorBody.dataset.filename;
    const ORIG_W       = parseInt(editorBody.dataset.originalWidth, 10);
    const ORIG_H       = parseInt(editorBody.dataset.originalHeight, 10);
    const COMPONENT_ID = editorBody.dataset.componentId || '';

    const DISPLAY_URL = `api/image-serve.php?file=${encodeURIComponent(FILENAME)}&w=1600&fit=max`;

    let cropper       = null;
    let activePreset  = null;
    let activeVariant = null;
    let presetsData   = [];

    // ── Load presets + crop state ─────────────────────────────────────────────

    async function loadState() {
        let url = `api/crops.php?image_id=${encodeURIComponent(IMAGE_ID)}`;
        if (COMPONENT_ID) url += `&component_id=${encodeURIComponent(COMPONENT_ID)}`;
        const res  = await fetch(url);
        const data = await res.json();
        presetsData = data.presets;
        renderPresetPanel();
    }

    // ── Render left panel ─────────────────────────────────────────────────────

    function renderPresetPanel() {
        const panel = document.getElementById('preset-list');
        panel.innerHTML = '';

        presetsData.forEach(preset => {
            const group = document.createElement('div');
            group.className = 'preset-group';
            group.dataset.presetId = preset.id;

            const header = document.createElement('div');
            header.className = 'preset-group-header';
            header.innerHTML = `
                <span>${esc(preset.label)}</span>
                <div class="preset-group-actions">
                    <button class="btn btn-sm btn-secondary btn-gen-all" title="Generate all variants">Generate All</button>
                </div>
            `;

            const list = document.createElement('ul');
            list.className = 'variant-list';

            preset.variants.forEach(variant => {
                const li = buildVariantItem(preset, variant);
                list.appendChild(li);
            });

            header.querySelector('.btn-gen-all').addEventListener('click', e => {
                e.stopPropagation();
                generateAll(preset.id);
            });

            group.appendChild(header);
            group.appendChild(list);
            panel.appendChild(group);
        });
    }

    function buildVariantItem(preset, variant) {
        const li = document.createElement('li');
        li.className = 'variant-item';
        li.dataset.presetId  = preset.id;
        li.dataset.variantId = variant.id;

        const badgeClass = variantBadge(variant);
        const badgeLabel = variant.generated_at ? 'Generated' : (variant.crop ? 'Saved' : 'Not set');

        li.innerHTML = `
            <div class="variant-info">
                <div class="variant-name">${esc(variant.label)}</div>
                <div class="variant-dims">${variant.width} × ${variant.height}px (×2: ${variant.width*2} × ${variant.height*2})</div>
            </div>
            <span class="badge ${badgeClass}">${badgeLabel}</span>
        `;

        li.addEventListener('click', () => selectVariant(preset, variant, li));
        return li;
    }

    function variantBadge(variant) {
        if (variant.generated_at) return 'badge-generated';
        if (variant.crop)         return 'badge-saved';
        return 'badge-unset';
    }

    // ── Select variant → init Cropper ─────────────────────────────────────────

    function selectVariant(preset, variant, liEl) {
        // Deactivate previous
        document.querySelectorAll('.variant-item.active').forEach(el => el.classList.remove('active'));
        liEl.classList.add('active');

        activePreset  = preset;
        activeVariant = variant;

        document.getElementById('crop-panel-empty').style.display  = 'none';
        document.getElementById('crop-panel-active').style.display = 'flex';

        document.getElementById('crop-info-label').textContent =
            `${preset.label} · ${variant.label} · ${variant.width}×${variant.height}px`;

        initCropper(variant);
    }

    function initCropper(variant) {
        const img = document.getElementById('crop-img');

        if (cropper) {
            cropper.destroy();
            cropper = null;
        }

        img.onload = () => {
            const displayW = img.naturalWidth;
            const displayH = img.naturalHeight;

            const scaleX = ORIG_W / displayW;
            const scaleY = ORIG_H / displayH;

            cropper = new Cropper(img, {
                aspectRatio: variant.width / variant.height,
                viewMode: 1,
                guides: true,
                center: true,
                background: false,
                autoCropArea: 0.85,
                ready() {
                    // Restore saved crop if available
                    if (variant.crop && variant.crop.w > 0) {
                        cropper.setData({
                            x: Math.round(variant.crop.x / scaleX),
                            y: Math.round(variant.crop.y / scaleY),
                            width:  Math.round(variant.crop.w / scaleX),
                            height: Math.round(variant.crop.h / scaleY),
                        });
                    }
                },
                crop() {
                    updateCoords(scaleX, scaleY);
                },
            });
        };

        // Set src after binding onload (re-trigger if same URL)
        if (img.src === location.origin + '/' + DISPLAY_URL || img.src === DISPLAY_URL) {
            img.dispatchEvent(new Event('load'));
        } else {
            img.src = DISPLAY_URL;
        }
    }

    function updateCoords(scaleX, scaleY) {
        if (!cropper) return;
        const d = cropper.getData(true);

        document.getElementById('coord-x').textContent = Math.round(d.x * scaleX);
        document.getElementById('coord-y').textContent = Math.round(d.y * scaleY);
        document.getElementById('coord-w').textContent = Math.round(d.width  * scaleX);
        document.getElementById('coord-h').textContent = Math.round(d.height * scaleY);
    }

    function getScaledCrop() {
        const img    = document.getElementById('crop-img');
        const scaleX = ORIG_W / img.naturalWidth;
        const scaleY = ORIG_H / img.naturalHeight;
        const d      = cropper.getData(true);

        return {
            x: Math.round(d.x * scaleX),
            y: Math.round(d.y * scaleY),
            w: Math.round(d.width  * scaleX),
            h: Math.round(d.height * scaleY),
        };
    }

    // ── Save crop ─────────────────────────────────────────────────────────────

    document.getElementById('btn-save-crop').addEventListener('click', async () => {
        if (!cropper || !activePreset || !activeVariant) return;

        const { x, y, w, h } = getScaledCrop();

        const res = await fetch('api/crops.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                image_id:   IMAGE_ID,
                preset_id:  activePreset.id,
                variant_id: activeVariant.id,
                crop_x: x, crop_y: y, crop_w: w, crop_h: h,
            }),
        });

        if (res.ok) {
            // Update local state + re-render badge
            activeVariant.crop = { x, y, w, h };
            activeVariant.generated_at = null;
            refreshVariantBadge(activePreset.id, activeVariant.id);
        } else {
            alert('Failed to save crop.');
        }
    });

    // ── Generate ──────────────────────────────────────────────────────────────

    async function generateAll(presetId) {
        const payload = { image_id: IMAGE_ID, preset_id: presetId };
        if (COMPONENT_ID) payload.component_id = COMPONENT_ID;

        const res  = await fetch('api/generate.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        const data = await res.json();

        if (data.error) {
            alert(data.error);
            return;
        }

        // Update local state for generated variants
        const preset = presetsData.find(p => p.id === presetId);
        if (preset) {
            data.generated.forEach(g => {
                const variant = preset.variants.find(v => v.id === g.variant_id);
                if (variant) {
                    variant.generated_at = new Date().toISOString();
                    variant.files        = g.files;
                    variant.files_exist  = true;
                    refreshVariantBadge(presetId, variant.id);
                }
            });
        }
    }

    // ── Open folder ───────────────────────────────────────────────────────────

    async function openFolder() {
        const payload = { image_id: IMAGE_ID };
        if (COMPONENT_ID) payload.component_id = COMPONENT_ID;

        const res = await fetch('api/open-folder.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        if (!res.ok) {
            const data = await res.json();
            alert(data.error || 'Could not open folder. Make sure the output folder is configured in Settings.');
        }
    }

    // ── Refresh single variant badge without full re-render ───────────────────

    function refreshVariantBadge(presetId, variantId) {
        const li = document.querySelector(
            `.variant-item[data-preset-id="${presetId}"][data-variant-id="${variantId}"]`
        );
        if (!li) return;

        const preset  = presetsData.find(p => p.id === presetId);
        const variant = preset?.variants.find(v => v.id === variantId);
        if (!variant) return;

        const badge = li.querySelector('.badge');
        badge.className = 'badge ' + variantBadge(variant);
        badge.textContent = variant.generated_at ? 'Generated' : (variant.crop ? 'Saved' : 'Not set');
    }

    // ── Helpers ───────────────────────────────────────────────────────────────

    function esc(str) {
        return String(str)
            .replace(/&/g,'&amp;').replace(/</g,'&lt;')
            .replace(/>/g,'&gt;').replace(/"/g,'&quot;');
    }

    // ── Init ──────────────────────────────────────────────────────────────────
    document.getElementById('btn-open-folder').addEventListener('click', openFolder);
    loadState();
})();

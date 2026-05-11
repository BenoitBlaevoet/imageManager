(() => {
    // ── Modal helper ──────────────────────────────────────────────────────────

    function showModal(id) {
        window.__modalZ = (window.__modalZ || 200) + 1;
        const el = document.getElementById(id);
        el.style.zIndex = window.__modalZ;
        el.style.display = 'flex';
    }

    // ── Tab navigation ────────────────────────────────────────────────────────

    function switchTab(tab) {
        const isImages = tab === 'images';
        document.getElementById('section-components').style.display = isImages ? 'none' : '';
        document.getElementById('section-images').style.display     = isImages ? '' : 'none';
        document.getElementById('tab-components').classList.toggle('active', !isImages);
        document.getElementById('tab-images').classList.toggle('active', isImages);

        if (isImages) {
            ImageLib.loadTags();
            ImageLib.loadLibraryImages();
        }
    }

    document.getElementById('tab-components').addEventListener('click', () => switchTab('components'));
    document.getElementById('tab-images').addEventListener('click', () => switchTab('images'));

    // ── Load components ───────────────────────────────────────────────────────

    async function loadComponents() {
        const res  = await fetch('api/components.php');
        const data = await res.json();
        renderComponentGrid(data.components);
    }
    window.loadComponentsCallback = loadComponents;

    function renderComponentGrid(components) {
        const grid = document.getElementById('component-grid');
        grid.innerHTML = '';

        if (components.length === 0) {
            grid.innerHTML = '<div class="loading-state">No components defined yet. Click "Manage Components" to add some.</div>';
            return;
        }

        components.forEach(c => grid.appendChild(buildComponentCard(c)));
    }

    function buildComponentCard(comp) {
        const card = document.createElement('div');
        card.className = 'component-card';
        card.dataset.id = comp.id;

        const hasImage    = comp.active_image !== null;
        const displayName = hasImage ? (comp.active_image.name || comp.active_image.original_name) : '';
        const statusText  = !hasImage
            ? 'No image'
            : comp.variants_generated + ' / ' + comp.variants_total + ' variants generated';
        const statusClass = !hasImage
            ? 'status-none'
            : (comp.variants_generated === comp.variants_total && comp.variants_total > 0 ? 'status-done' : 'status-partial');

        card.innerHTML = `
            <div class="component-card-thumb">
                ${hasImage
                    ? `<img src="api/${esc(comp.active_image.thumb_url)}" alt="${esc(displayName)}">`
                    : `<div class="component-card-no-image"><svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg><span>No image</span></div>`
                }
            </div>
            <div class="component-card-body">
                <div class="component-card-name">${esc(comp.label)}</div>
                ${comp.description ? `<div class="component-card-desc">${esc(comp.description)}</div>` : ''}
                <div class="component-card-meta">
                    <span class="component-status ${statusClass}">${statusText}</span>
                    ${comp.image_count > 0 ? `<button class="btn-link btn-image-count" data-id="${esc(comp.id)}" data-label="${esc(comp.label)}">${comp.image_count} image${comp.image_count > 1 ? 's' : ''}</button>` : ''}
                </div>
                <div class="component-card-actions">
                    <button class="btn btn-secondary btn-sm btn-pick-img" data-id="${esc(comp.id)}" data-label="${esc(comp.label)}">
                        ${hasImage ? 'Change Image' : 'Add Image'}
                    </button>
                    ${hasImage ? `<a href="editor.php?component_id=${esc(comp.id)}&image_id=${esc(comp.active_image.id)}" class="btn btn-primary btn-sm">Edit Crops</a>` : ''}
                    ${hasImage ? `<button class="btn btn-ghost btn-sm btn-open-folder" title="Open output folder"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg></button>` : ''}
                </div>
            </div>
        `;

        card.querySelector('.btn-pick-img').addEventListener('click', () =>
            openImagePickerModal(comp.id, comp.label)
        );

        card.querySelector('.btn-open-folder')?.addEventListener('click', async () => {
            const res = await fetch('api/open-folder.php', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ component_id: comp.id, image_id: comp.active_image.id }),
            });
            if (!res.ok) {
                const data = await res.json();
                alert(data.error || 'Could not open folder.');
            }
        });

        const countBtn = card.querySelector('.btn-image-count');
        if (countBtn) {
            countBtn.addEventListener('click', () => openImageBrowser(comp.id, comp.label));
        }

        return card;
    }

    // ── Image picker modal ────────────────────────────────────────────────────

    let pickerComponentId    = null;
    let pickerComponentLabel = null;
    let pickerActiveTagIds   = new Set();
    let pickerSearchTimer    = null;

    function openImagePickerModal(componentId, componentLabel) {
        pickerComponentId    = componentId;
        pickerComponentLabel = componentLabel;
        pickerActiveTagIds   = new Set();

        document.getElementById('image-picker-title').textContent = componentLabel + ' — Add Image';
        showModal('image-picker-modal');
        loadPickerImages();
        loadPickerTagFilter();
    }

    function closeImagePickerModal() {
        pickerComponentId    = null;
        pickerComponentLabel = null;
        document.getElementById('image-picker-modal').style.display = 'none';
    }

    async function loadPickerTagFilter() {
        const tags = ImageLib.getTags().length
            ? ImageLib.getTags()
            : (await (await fetch('api/tags.php')).json()).tags || [];

        const bar = document.getElementById('picker-tag-filters');
        bar.innerHTML = '';

        tags.forEach(tag => {
            const chip = document.createElement('button');
            chip.type = 'button';
            chip.className = 'tag-filter-chip' + (pickerActiveTagIds.has(tag.id) ? ' active' : '');
            chip.textContent = tag.name;
            chip.addEventListener('click', () => {
                if (pickerActiveTagIds.has(tag.id)) {
                    pickerActiveTagIds.delete(tag.id);
                } else {
                    pickerActiveTagIds.add(tag.id);
                }
                chip.classList.toggle('active');
                loadPickerImages();
            });
            bar.appendChild(chip);
        });
    }

    async function loadPickerImages() {
        const search = document.getElementById('picker-search')?.value?.trim() || '';
        const tagIds = [...pickerActiveTagIds];

        const params = new URLSearchParams();
        if (search) params.append('search', search);
        tagIds.forEach(id => params.append('tags[]', id));
        const url = 'api/images.php' + (params.toString() ? '?' + params.toString() : '');

        const grid = document.getElementById('picker-image-grid');
        grid.innerHTML = '<div class="loading-state">Loading…</div>';

        const res    = await fetch(url);
        const images = await res.json();

        grid.innerHTML = '';
        if (!images.length) {
            grid.innerHTML = '<div class="loading-state">No images found.</div>';
            return;
        }

        images.forEach(img => {
            const displayName = img.name || img.original_name;
            const tagsHtml = (img.tags || []).map(t => `<span class="tag-chip">${esc(t.name)}</span>`).join('');
            const card = document.createElement('div');
            card.className = 'browser-image-card';
            card.innerHTML = `
                <img src="api/${esc(img.thumb_url)}" alt="${esc(displayName)}" loading="lazy">
                <div class="browser-image-info">
                    <div class="browser-image-name">${esc(displayName)}</div>
                    <div class="browser-image-dims">${img.width} × ${img.height}px</div>
                    ${tagsHtml ? `<div style="margin-top:.25rem">${tagsHtml}</div>` : ''}
                </div>
                <div class="browser-image-actions">
                    <button class="btn btn-primary btn-sm btn-picker-add" data-id="${esc(img.id)}">Add</button>
                </div>
            `;
            card.querySelector('.btn-picker-add').addEventListener('click', async () => {
                await fetch('api/component-images.php', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ component_id: pickerComponentId, image_id: img.id }),
                });
                closeImagePickerModal();
                loadComponents();
            });
            grid.appendChild(card);
        });
    }

    document.getElementById('image-picker-close')?.addEventListener('click', closeImagePickerModal);
    document.getElementById('btn-picker-cancel')?.addEventListener('click', closeImagePickerModal);
    document.getElementById('btn-picker-upload')?.addEventListener('click', () => {
        openUploadModal({
            componentId:    pickerComponentId,
            componentLabel: pickerComponentLabel,
            onSuccess: () => {
                loadPickerImages();
                loadComponents();
            },
        });
    });
    document.getElementById('picker-search')?.addEventListener('input', () => {
        clearTimeout(pickerSearchTimer);
        pickerSearchTimer = setTimeout(loadPickerImages, 300);
    });

    // ── Image browser modal (component image list) ────────────────────────────

    let imageBrowserComponentId = null;

    function openImageBrowser(componentId, componentLabel) {
        imageBrowserComponentId = componentId;
        document.getElementById('image-browser-title').textContent = componentLabel + ' — Assigned Images';
        showModal('image-browser-modal');
        loadImageBrowser(componentId);
    }

    function closeImageBrowser() {
        imageBrowserComponentId = null;
        document.getElementById('image-browser-modal').style.display = 'none';
    }

    async function loadImageBrowser(componentId) {
        const grid = document.getElementById('image-browser-grid');
        grid.innerHTML = '<div class="loading-state">Loading…</div>';

        const res  = await fetch(`api/component-images.php?component_id=${encodeURIComponent(componentId)}`);
        const data = await res.json();

        grid.innerHTML = '';
        if (!data.images.length) {
            grid.innerHTML = '<div class="loading-state">No images assigned yet.</div>';
            return;
        }

        data.images.forEach(img => {
            const displayName = img.name || img.original_name;
            const tagsHtml = (img.tags || []).map(t => `<span class="tag-chip">${esc(t.name)}</span>`).join('');
            const card = document.createElement('div');
            card.className = 'browser-image-card' + (img.is_active ? ' is-active' : '');
            card.innerHTML = `
                <img src="api/${esc(img.thumb_url)}" alt="${esc(displayName)}">
                <div class="browser-image-info">
                    <div class="browser-image-name">${esc(displayName)}</div>
                    <div class="browser-image-dims">${img.width} × ${img.height}px</div>
                    ${tagsHtml ? `<div style="margin-top:.25rem">${tagsHtml}</div>` : ''}
                </div>
                <div class="browser-image-actions">
                    ${!img.is_active
                        ? `<button class="btn btn-primary btn-sm btn-set-active" data-assignment="${img.assignment_id}">Set Active</button>`
                        : '<span class="badge-active">Active</span>'}
                    <button class="btn btn-ghost btn-sm btn-remove-img" data-assignment="${img.assignment_id}">Remove</button>
                </div>
            `;

            const setActive = card.querySelector('.btn-set-active');
            if (setActive) {
                setActive.addEventListener('click', async () => {
                    await fetch('api/component-images.php', {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ assignment_id: img.assignment_id }),
                    });
                    loadImageBrowser(componentId);
                    loadComponents();
                });
            }

            card.querySelector('.btn-remove-img').addEventListener('click', async () => {
                if (!confirm('Remove this image from the component?')) return;
                await fetch(`api/component-images.php?id=${img.assignment_id}`, { method: 'DELETE' });
                loadImageBrowser(componentId);
                loadComponents();
            });

            grid.appendChild(card);
        });
    }

    document.getElementById('image-browser-close')?.addEventListener('click', closeImageBrowser);

    // ── Settings modal ────────────────────────────────────────────────────────

    async function openSettings() {
        const res    = await fetch('api/config.php');
        const config = await res.json();
        document.getElementById('output-folder').value  = config.output_folder || '';
        document.getElementById('default-format').value = config.default_format || 'webp';
        showModal('settings-modal');
    }

    function closeSettings() { document.getElementById('settings-modal').style.display = 'none'; }

    document.getElementById('btn-settings').addEventListener('click', openSettings);
    document.getElementById('settings-close').addEventListener('click', closeSettings);

    document.getElementById('btn-save-settings').addEventListener('click', async () => {
        const folder = document.getElementById('output-folder').value.trim();
        const format = document.getElementById('default-format').value;
        const res    = await fetch('api/config.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ output_folder: folder, default_format: format }),
        });
        const data = await res.json();
        if (data.error) { alert(data.error); return; }
        closeSettings();
    });

    // ── Presets Manager ───────────────────────────────────────────────────────

    const presetsModal = document.getElementById('presets-modal');
    let presetsWorkingCopy = [];
    let selectedPresetId   = null;

    async function openPresetsModal() {
        showModal('presets-modal');
        const res  = await fetch('api/presets.php');
        const data = await res.json();
        presetsWorkingCopy = JSON.parse(JSON.stringify(data.presets));
        selectedPresetId   = null;
        renderPresetTree();
        clearPresetForm();
    }

    function closePresetsModal() { presetsModal.style.display = 'none'; }

    function renderPresetTree() {
        const tree = document.getElementById('preset-tree');
        if (presetsWorkingCopy.length === 0) {
            tree.innerHTML = '<div style="padding:1rem;text-align:center;color:var(--text-muted)">No presets</div>';
            return;
        }
        tree.innerHTML = '';
        presetsWorkingCopy.forEach((preset, idx) => {
            const isActive = selectedPresetId === preset.id;
            const item = document.createElement('div');
            item.className = 'preset-tree-item';
            item.innerHTML = `
                <div class="preset-item-header ${isActive ? 'active' : ''}">
                    <span class="preset-chevron">▶</span>
                    <span class="preset-item-label">${esc(preset.label)}</span>
                    <div class="preset-item-actions">
                        <button type="button" class="btn btn-ghost btn-sm btn-delete-preset">✕</button>
                    </div>
                </div>
            `;
            item.querySelector('.preset-item-header').addEventListener('click', () => {
                selectedPresetId = preset.id;
                renderPresetTree();
                showPresetForm(idx);
            });
            item.querySelector('.btn-delete-preset').addEventListener('click', e => {
                e.stopPropagation();
                if (confirm(`Delete preset "${preset.label}"?`)) {
                    presetsWorkingCopy.splice(idx, 1);
                    selectedPresetId = null;
                    renderPresetTree();
                    clearPresetForm();
                }
            });
            tree.appendChild(item);
        });
    }

    function showPresetForm(presetIdx) {
        const preset = presetsWorkingCopy[presetIdx];
        document.getElementById('preset-form-empty').style.display = 'none';
        const form = document.getElementById('preset-edit-form');
        form.style.display = 'flex';
        document.getElementById('form-preset-id').value     = preset.id;
        document.getElementById('form-preset-label').value  = preset.label;
        document.getElementById('form-preset-format').value = preset.format || '';

        document.getElementById('variants-list').innerHTML = '';
        (preset.variants || []).forEach((v, vIdx) => {
            document.getElementById('variants-list').appendChild(buildVariantFormElement(presetIdx, vIdx, v));
        });

        document.getElementById('btn-add-variant').onclick = () => {
            preset.variants.push({ id: 'variant-' + Date.now(), label: 'New Variant', width: 400, height: 400 });
            showPresetForm(presetIdx);
        };
        document.getElementById('btn-delete-preset').onclick = () => {
            if (confirm(`Delete "${preset.label}"?`)) {
                presetsWorkingCopy.splice(presetIdx, 1);
                selectedPresetId = null;
                renderPresetTree();
                clearPresetForm();
            }
        };
        document.getElementById('form-preset-id').onchange     = e => { preset.id = e.target.value; renderPresetTree(); };
        document.getElementById('form-preset-label').onchange  = e => { preset.label = e.target.value; renderPresetTree(); };
        document.getElementById('form-preset-format').onchange = e => { preset.format = e.target.value || undefined; };
    }

    function buildVariantFormElement(presetIdx, variantIdx, variant) {
        const div = document.createElement('div');
        div.className = 'variant-form';
        div.innerHTML = `
            <div class="form-row"><label class="field-label">ID</label><input type="text" class="field-input variant-id" value="${esc(variant.id)}"></div>
            <div class="form-row"><label class="field-label">Name</label><input type="text" class="field-input variant-label" value="${esc(variant.label)}"></div>
            <div class="form-row form-2col">
                <div><label class="field-label">Width</label><input type="number" class="field-input variant-width" value="${variant.width}" min="1"></div>
                <div><label class="field-label">Height</label><input type="number" class="field-input variant-height" value="${variant.height}" min="1"></div>
            </div>
            <div class="form-row"><label class="field-label">Format (optional)</label>
                <select class="field-input variant-format">
                    <option value="">Inherit</option>
                    <option value="webp" ${variant.format === 'webp' ? 'selected' : ''}>WebP</option>
                    <option value="jpg"  ${variant.format === 'jpg'  ? 'selected' : ''}>JPEG</option>
                    <option value="png"  ${variant.format === 'png'  ? 'selected' : ''}>PNG</option>
                </select>
            </div>
            <button type="button" class="btn btn-danger btn-sm variant-form-remove">Delete Variant</button>
        `;
        div.querySelector('.variant-id').onchange     = e => { variant.id = e.target.value; };
        div.querySelector('.variant-label').onchange  = e => { variant.label = e.target.value; renderPresetTree(); };
        div.querySelector('.variant-width').onchange  = e => { variant.width  = parseInt(e.target.value, 10) || 0; };
        div.querySelector('.variant-height').onchange = e => { variant.height = parseInt(e.target.value, 10) || 0; };
        div.querySelector('.variant-format').onchange = e => { variant.format = e.target.value || undefined; };
        div.querySelector('.variant-form-remove').addEventListener('click', () => {
            if (confirm(`Delete variant "${variant.label}"?`)) {
                presetsWorkingCopy[presetIdx].variants.splice(variantIdx, 1);
                showPresetForm(presetIdx);
            }
        });
        return div;
    }

    function clearPresetForm() {
        document.getElementById('preset-form-empty').style.display = 'flex';
        document.getElementById('preset-edit-form').style.display  = 'none';
    }

    document.getElementById('btn-presets').addEventListener('click', openPresetsModal);
    document.getElementById('presets-close').addEventListener('click', closePresetsModal);
    document.getElementById('btn-close-presets').addEventListener('click', closePresetsModal);

    document.getElementById('btn-add-preset').addEventListener('click', () => {
        const p = { id: 'new-preset-' + Date.now(), label: 'New Preset', variants: [{ id: 'variant-1', label: 'Variant 1', width: 400, height: 300 }] };
        presetsWorkingCopy.push(p);
        selectedPresetId = p.id;
        renderPresetTree();
        showPresetForm(presetsWorkingCopy.length - 1);
    });

    document.getElementById('btn-save-presets').addEventListener('click', async () => {
        const res  = await fetch('api/presets.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ presets: presetsWorkingCopy }),
        });
        const data = await res.json();
        if (data.error) { alert('Error:\n' + (data.errors ? data.errors.join('\n') : data.error)); return; }
        closePresetsModal();
    });

    // ── Components Manager ────────────────────────────────────────────────────

    const componentsModal = document.getElementById('components-modal');
    let componentsWorkingCopy = [];
    let selectedComponentId   = null;
    let availablePresets      = [];

    async function openComponentsModal() {
        showModal('components-modal');
        const [cRes, pRes] = await Promise.all([fetch('api/components.php'), fetch('api/presets.php')]);
        const cData = await cRes.json();
        const pData = await pRes.json();

        componentsWorkingCopy = JSON.parse(JSON.stringify(
            cData.components.map(c => ({ id: c.id, label: c.label, description: c.description || '', presets: c.presets }))
        ));
        availablePresets    = pData.presets;
        selectedComponentId = null;
        renderComponentTree();
        clearComponentForm();
    }

    function closeComponentsModal() { componentsModal.style.display = 'none'; }

    function renderComponentTree() {
        const tree = document.getElementById('component-tree');
        if (componentsWorkingCopy.length === 0) {
            tree.innerHTML = '<div style="padding:1rem;text-align:center;color:var(--text-muted)">No components</div>';
            return;
        }
        tree.innerHTML = '';
        componentsWorkingCopy.forEach((comp, idx) => {
            const isActive = selectedComponentId === comp.id;
            const item = document.createElement('div');
            item.className = 'preset-tree-item';
            item.innerHTML = `
                <div class="preset-item-header ${isActive ? 'active' : ''}">
                    <span class="preset-item-label">${esc(comp.label)}</span>
                    <div class="preset-item-actions">
                        <button type="button" class="btn btn-ghost btn-sm btn-delete-component">✕</button>
                    </div>
                </div>
            `;
            item.querySelector('.preset-item-header').addEventListener('click', () => {
                selectedComponentId = comp.id;
                renderComponentTree();
                showComponentForm(idx);
            });
            item.querySelector('.btn-delete-component').addEventListener('click', e => {
                e.stopPropagation();
                if (confirm(`Delete component "${comp.label}"?`)) {
                    componentsWorkingCopy.splice(idx, 1);
                    selectedComponentId = null;
                    renderComponentTree();
                    clearComponentForm();
                }
            });
            tree.appendChild(item);
        });
    }

    function showComponentForm(compIdx) {
        const comp = componentsWorkingCopy[compIdx];
        document.getElementById('component-form-empty').style.display = 'none';
        document.getElementById('component-edit-form').style.display  = 'flex';
        document.getElementById('form-component-id').value          = comp.id;
        document.getElementById('form-component-label').value       = comp.label;
        document.getElementById('form-component-description').value = comp.description || '';

        const checkboxes = document.getElementById('component-preset-checkboxes');
        checkboxes.innerHTML = '';
        availablePresets.forEach(preset => {
            const checked = comp.presets.includes(preset.id) ? 'checked' : '';
            const label   = document.createElement('label');
            label.className = 'preset-checkbox-item';
            label.innerHTML = `<input type="checkbox" value="${esc(preset.id)}" ${checked}> ${esc(preset.label)}`;
            checkboxes.appendChild(label);
        });

        document.getElementById('form-component-id').onchange          = e => { comp.id = e.target.value; renderComponentTree(); };
        document.getElementById('form-component-label').onchange       = e => { comp.label = e.target.value; renderComponentTree(); };
        document.getElementById('form-component-description').onchange = e => { comp.description = e.target.value; };
        checkboxes.onchange = () => {
            comp.presets = [...checkboxes.querySelectorAll('input:checked')].map(cb => cb.value);
        };
        document.getElementById('btn-delete-component').onclick = () => {
            if (confirm(`Delete "${comp.label}"?`)) {
                componentsWorkingCopy.splice(compIdx, 1);
                selectedComponentId = null;
                renderComponentTree();
                clearComponentForm();
            }
        };
    }

    function clearComponentForm() {
        document.getElementById('component-form-empty').style.display = 'flex';
        document.getElementById('component-edit-form').style.display  = 'none';
    }

    document.getElementById('btn-components').addEventListener('click', openComponentsModal);
    document.getElementById('components-close').addEventListener('click', closeComponentsModal);
    document.getElementById('btn-close-components').addEventListener('click', closeComponentsModal);

    document.getElementById('btn-add-component').addEventListener('click', () => {
        const c = { id: 'component-' + Date.now(), label: 'New Component', description: '', presets: [] };
        componentsWorkingCopy.push(c);
        selectedComponentId = c.id;
        renderComponentTree();
        showComponentForm(componentsWorkingCopy.length - 1);
    });

    document.getElementById('btn-save-components').addEventListener('click', async () => {
        const res  = await fetch('api/components.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ components: componentsWorkingCopy }),
        });
        const data = await res.json();
        if (data.error) { alert('Error:\n' + (data.errors ? data.errors.join('\n') : data.error)); return; }
        closeComponentsModal();
        loadComponents();
    });

    // ── Helpers ───────────────────────────────────────────────────────────────

    function esc(str) {
        return String(str)
            .replace(/&/g, '&amp;').replace(/</g, '&lt;')
            .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    }

    // ── Init ──────────────────────────────────────────────────────────────────
    loadComponents();
})();

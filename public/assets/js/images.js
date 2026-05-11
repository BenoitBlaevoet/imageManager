(() => {
    let allTags = [];
    let activeSectionTagIds = new Set();
    let pendingUploadFile   = null;
    let uploadOpts          = {};

    // ── Shared modal z-index ──────────────────────────────────────────────────

    function showModal(id) {
        window.__modalZ = (window.__modalZ || 200) + 1;
        const el = document.getElementById(id);
        el.style.zIndex = window.__modalZ;
        el.style.display = 'flex';
    }

    // ── Tags ──────────────────────────────────────────────────────────────────

    async function loadTags() {
        const res  = await fetch('api/tags.php');
        const data = await res.json();
        allTags = data.tags || [];
        renderSectionTagFilter();
        return allTags;
    }

    function renderSectionTagFilter() {
        const bar = document.getElementById('images-tag-filter');
        if (!bar) return;
        bar.innerHTML = '';

        if (!allTags.length) return;

        allTags.forEach(tag => {
            const chip = document.createElement('button');
            chip.type = 'button';
            chip.className = 'tag-filter-chip' + (activeSectionTagIds.has(tag.id) ? ' active' : '');
            chip.textContent = tag.name + (tag.image_count ? ' (' + tag.image_count + ')' : '');
            chip.addEventListener('click', () => {
                if (activeSectionTagIds.has(tag.id)) {
                    activeSectionTagIds.delete(tag.id);
                } else {
                    activeSectionTagIds.add(tag.id);
                }
                chip.classList.toggle('active');
                loadLibraryImages();
            });
            bar.appendChild(chip);
        });
    }

    // ── Image library ─────────────────────────────────────────────────────────

    async function loadLibraryImages() {
        const search = document.getElementById('images-search')?.value?.trim() || '';
        const tagIds = [...activeSectionTagIds];

        const params = new URLSearchParams();
        if (search) params.append('search', search);
        tagIds.forEach(id => params.append('tags[]', id));
        const url = 'api/images.php' + (params.toString() ? '?' + params.toString() : '');

        const grid = document.getElementById('image-library-grid');
        if (grid) grid.innerHTML = '<div class="loading-state">Loading…</div>';

        const res    = await fetch(url);
        const images = await res.json();
        renderLibraryGrid(images);
    }

    function renderLibraryGrid(images) {
        const grid = document.getElementById('image-library-grid');
        if (!grid) return;
        grid.innerHTML = '';

        if (!images.length) {
            grid.innerHTML = '<div class="loading-state">No images found.</div>';
            return;
        }

        images.forEach(img => grid.appendChild(buildLibraryCard(img)));
    }

    function buildLibraryCard(img) {
        const card = document.createElement('div');
        card.className = 'library-image-card';

        const displayName = img.name || img.original_name;
        const showOriginal = img.name && img.name !== img.original_name;
        const tagsHtml = (img.tags || []).map(t => `<span class="tag-chip">${esc(t.name)}</span>`).join('');

        card.innerHTML = `
            <img src="api/${esc(img.thumb_url)}" alt="${esc(displayName)}" loading="lazy">
            <div class="library-image-info">
                <div class="library-image-name">${esc(displayName)}</div>
                ${showOriginal ? `<div class="library-image-original">${esc(img.original_name)}</div>` : ''}
                <div class="library-image-dims">${img.width} × ${img.height}px</div>
                ${tagsHtml ? `<div class="library-image-tags">${tagsHtml}</div>` : ''}
            </div>
            <div class="library-image-actions">
                <button class="btn btn-danger btn-sm btn-lib-delete" data-id="${esc(img.id)}"
                    data-name="${esc(displayName)}">Delete</button>
            </div>
        `;

        card.querySelector('.btn-lib-delete').addEventListener('click', async () => {
            const name = card.querySelector('.btn-lib-delete').dataset.name;
            if (!confirm(`Delete "${name}"? This cannot be undone.`)) return;
            const res = await fetch(`api/images.php?id=${encodeURIComponent(img.id)}`, { method: 'DELETE' });
            if (res.ok) {
                card.remove();
                loadTags(); // refresh counts
                if (window.loadComponentsCallback) window.loadComponentsCallback();
            }
        });

        return card;
    }

    // ── Search debounce ───────────────────────────────────────────────────────

    let searchTimer = null;
    function onSearchInput() {
        clearTimeout(searchTimer);
        searchTimer = setTimeout(loadLibraryImages, 300);
    }

    // ── Upload modal ──────────────────────────────────────────────────────────

    function openUploadModal(opts) {
        uploadOpts          = opts || {};
        pendingUploadFile   = null;

        const ctxLabel = document.getElementById('upload-context-label');
        if (uploadOpts.componentLabel) {
            ctxLabel.textContent = 'Adding to: ' + uploadOpts.componentLabel;
            ctxLabel.style.display = '';
        } else {
            ctxLabel.style.display = 'none';
        }

        showUploadStep(1);
        showModal('upload-modal');
    }
    window.openUploadModal = openUploadModal;

    function closeUploadModal() {
        document.getElementById('upload-modal').style.display = 'none';
        pendingUploadFile = null;
        uploadOpts = {};
    }

    function showUploadStep(step) {
        document.getElementById('upload-step-1').style.display    = step === 1 ? '' : 'none';
        document.getElementById('upload-step-2').style.display    = step === 2 ? '' : 'none';
        document.getElementById('upload-footer').style.display    = step === 2 ? 'flex' : 'none';
        document.getElementById('upload-progress').style.display  = 'none';
    }

    async function onFileSelected(file) {
        if (!file) return;
        pendingUploadFile = file;

        document.getElementById('upload-name').value =
            file.name.replace(/\.[^.]+$/, '').replace(/[-_]+/g, ' ').trim();

        if (!allTags.length) await loadTags();
        renderUploadTagCheckboxes();
        showUploadStep(2);
    }

    function renderUploadTagCheckboxes() {
        const list = document.getElementById('upload-tags-list');
        list.innerHTML = '';
        allTags.forEach(tag => {
            const lbl = document.createElement('label');
            lbl.className = 'tag-checkbox-item';
            lbl.innerHTML = `<input type="checkbox" value="${tag.id}"> ${esc(tag.name)}`;
            list.appendChild(lbl);
        });
    }

    async function addTagInline() {
        const input = document.getElementById('upload-new-tag');
        const name  = input.value.trim();
        if (!name) return;

        const res  = await fetch('api/tags.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name }),
        });
        const data = await res.json();
        if (data.error) { alert(data.error); return; }

        // Merge into allTags
        const existing = allTags.find(t => t.id === data.id);
        if (!existing) allTags.push({ id: data.id, name: data.name, image_count: 0 });

        input.value = '';
        renderUploadTagCheckboxes();
        renderSectionTagFilter();

        // Auto-check the new tag
        const boxes = document.getElementById('upload-tags-list').querySelectorAll('input');
        const last  = boxes[boxes.length - 1];
        if (last && last.value == data.id) last.checked = true;
    }

    async function doUpload() {
        if (!pendingUploadFile) return;

        const name   = document.getElementById('upload-name').value.trim();
        const tagIds = [...document.getElementById('upload-tags-list').querySelectorAll('input:checked')]
            .map(cb => cb.value);

        document.getElementById('upload-step-2').style.display  = 'none';
        document.getElementById('upload-footer').style.display  = 'none';
        document.getElementById('upload-progress').style.display = '';
        const fill  = document.getElementById('progress-fill');
        const label = document.getElementById('progress-label');
        fill.style.width = '0%';

        const formData = new FormData();
        formData.append('image', pendingUploadFile);
        if (name) formData.append('name', name);
        if (tagIds.length) formData.append('tags', JSON.stringify(tagIds));
        if (uploadOpts.componentId) formData.append('component_id', uploadOpts.componentId);

        let pct = 0;
        const timer = setInterval(() => { pct = Math.min(pct + 10, 85); fill.style.width = pct + '%'; }, 120);

        try {
            const res  = await fetch('api/upload.php', { method: 'POST', body: formData });
            const data = await res.json();
            clearInterval(timer);

            if (data.error) { alert(data.error); showUploadStep(1); return; }

            fill.style.width = '100%';
            label.textContent = 'Done!';

            setTimeout(() => {
                closeUploadModal();
                loadTags();
                if (document.getElementById('section-images')?.style.display !== 'none') {
                    loadLibraryImages();
                }
                if (uploadOpts.onSuccess) uploadOpts.onSuccess(data);
            }, 400);
        } catch {
            clearInterval(timer);
            alert('Upload failed.');
            showUploadStep(1);
        }
    }

    // ── Tag manager modal ─────────────────────────────────────────────────────

    async function openTagManager() {
        await loadTags();
        renderTagManagerList();
        showModal('tag-manager-modal');
    }

    function renderTagManagerList() {
        const list = document.getElementById('tag-manager-list');
        if (!list) return;
        list.innerHTML = '';

        if (!allTags.length) {
            list.innerHTML = '<div class="loading-state" style="padding:1rem 0">No tags yet.</div>';
            return;
        }

        allTags.forEach(tag => {
            const row = document.createElement('div');
            row.className = 'tag-manager-row';
            row.innerHTML = `
                <span class="tag-chip">${esc(tag.name)}</span>
                <span class="tag-manager-count">${tag.image_count} image${tag.image_count !== 1 ? 's' : ''}</span>
                <button class="btn btn-danger btn-sm" data-id="${tag.id}">Delete</button>
            `;
            row.querySelector('button').addEventListener('click', async () => {
                if (!confirm(`Delete tag "${tag.name}"? It will be removed from all images.`)) return;
                await fetch(`api/tags.php?id=${tag.id}`, { method: 'DELETE' });
                allTags = allTags.filter(t => t.id !== tag.id);
                activeSectionTagIds.delete(tag.id);
                renderTagManagerList();
                renderSectionTagFilter();
                if (document.getElementById('section-images')?.style.display !== 'none') {
                    loadLibraryImages();
                }
            });
            list.appendChild(row);
        });
    }

    async function createTagFromManager() {
        const input = document.getElementById('new-tag-input');
        const name  = input.value.trim();
        if (!name) return;

        const res  = await fetch('api/tags.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name }),
        });
        const data = await res.json();
        if (data.error) { alert(data.error); return; }

        if (!allTags.find(t => t.id === data.id)) {
            allTags.push({ id: data.id, name: data.name, image_count: 0 });
        }
        input.value = '';
        renderTagManagerList();
        renderSectionTagFilter();
    }

    // ── Init ──────────────────────────────────────────────────────────────────

    function init() {
        // Search
        document.getElementById('images-search')?.addEventListener('input', onSearchInput);

        // Upload button (Images section)
        document.getElementById('btn-upload-image')?.addEventListener('click', () => openUploadModal());

        // Upload modal close
        document.getElementById('upload-close')?.addEventListener('click', closeUploadModal);
        document.getElementById('upload-modal')?.addEventListener('click', e => {
            if (e.target.id === 'upload-modal') closeUploadModal();
        });

        // Drop zone
        const dz = document.getElementById('drop-zone');
        const fi = document.getElementById('file-input');
        if (dz) {
            dz.addEventListener('dragover', e => { e.preventDefault(); dz.classList.add('drag-over'); });
            dz.addEventListener('dragleave', () => dz.classList.remove('drag-over'));
            dz.addEventListener('drop', e => {
                e.preventDefault();
                dz.classList.remove('drag-over');
                onFileSelected(e.dataTransfer.files[0]);
            });
        }
        if (fi) fi.addEventListener('change', () => onFileSelected(fi.files[0]));

        // Upload form
        document.getElementById('btn-upload-back')?.addEventListener('click', () => showUploadStep(1));
        document.getElementById('btn-do-upload')?.addEventListener('click', doUpload);
        document.getElementById('btn-add-tag-upload')?.addEventListener('click', addTagInline);
        document.getElementById('upload-new-tag')?.addEventListener('keydown', e => {
            if (e.key === 'Enter') { e.preventDefault(); addTagInline(); }
        });

        // Tag manager
        document.getElementById('btn-manage-tags')?.addEventListener('click', openTagManager);
        document.getElementById('tag-manager-close')?.addEventListener('click', () => {
            document.getElementById('tag-manager-modal').style.display = 'none';
        });
        document.getElementById('tag-manager-modal')?.addEventListener('click', e => {
            if (e.target.id === 'tag-manager-modal')
                document.getElementById('tag-manager-modal').style.display = 'none';
        });
        document.getElementById('btn-create-tag')?.addEventListener('click', createTagFromManager);
        document.getElementById('new-tag-input')?.addEventListener('keydown', e => {
            if (e.key === 'Enter') { e.preventDefault(); createTagFromManager(); }
        });
    }

    // ── Helpers ───────────────────────────────────────────────────────────────

    function esc(str) {
        return String(str)
            .replace(/&/g, '&amp;').replace(/</g, '&lt;')
            .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    }

    // ── Expose ────────────────────────────────────────────────────────────────

    window.ImageLib = {
        init,
        loadTags,
        loadLibraryImages,
        getTags: () => allTags,
    };

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();

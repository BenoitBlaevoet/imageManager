# ImageManager

Component-based image management and generation tool. Upload images once, name and tag them, crop interactively per preset, generate 1x/2x output variants automatically.

## What it does

1. **Upload** images → stored in `storage/source/`, metadata in SQLite
2. **Name & tag** images for search and organisation
3. **Organize** images into *Components* (e.g. `partner-banner`, `product-hero`)
4. **Define** output *Presets* with variants (e.g. `homepage-banner` → mobile 375×200, large 1440×600)
5. **Crop** each image per variant using an interactive editor (Cropper.js)
6. **Generate** 1x and 2x output files → saved under `{component}/{image-name}/`

## Stack

- PHP 8.2+, [league/glide](https://glide.thephpleague.com/) `^3.0`
- SQLite (auto-migrated, no setup needed)
- Vanilla JS + [Cropper.js](https://fengyuanchen.github.io/cropperjs/)
- No framework

## Setup

```bash
git clone <repo>
cd ImageManager
composer install
php -S localhost:8000 -t public
```

Open http://localhost:8000.

On first request, the app auto-creates all required directories (`data/`, `storage/source/`, `storage/cache/`, `output/`) and sets the output folder to `{project}/output/`. No manual setup needed.

## Configuration

Settings are stored in `config/app-config.json` (gitignored, auto-created from `app-config.example.json`).

| Key | Default | Description |
|-----|---------|-------------|
| `output_folder` | `{project}/output` | Absolute path where generated images are written |
| `default_format` | `webp` | Output format when not specified by preset (`webp`, `jpg`, `png`) |

Change via the **Settings** button in the UI, or edit `config/app-config.json` directly.

## Project structure

```
├── config/
│   ├── app-config.json         # Machine-specific config (gitignored)
│   ├── app-config.example.json # Template
│   ├── components.json         # Component definitions
│   └── presets.json            # Output presets and variants
├── data/
│   └── imagemanager.db         # SQLite DB (auto-created, gitignored)
├── output/                     # Generated images (gitignored)
├── public/
│   ├── index.php               # Main UI (Components tab + Images tab)
│   ├── editor.php              # Crop editor
│   └── api/                   # REST API endpoints
├── src/                        # PHP application classes (App\ namespace)
│   ├── bootstrap.php           # Auto-setup (runs on every request via Composer)
│   ├── helpers.php             # Shared functions (slugify)
│   ├── Database.php            # SQLite singleton + auto-migration
│   ├── TagRepository.php       # Tag CRUD + image-tag pivot
│   ├── ImageRepository.php     # Image CRUD with search + tag filter
│   ├── GlideServer.php         # Glide image server factory
│   └── ...
├── storage/
│   ├── source/                 # Uploaded originals (gitignored)
│   └── cache/                  # Glide transform cache (gitignored)
└── vendor/
```

## UI overview

### Components tab (default)

- Grid of all defined components, each showing its active image thumbnail and generation status
- **Add Image / Change Image** → opens the image picker modal: browse the full library, search by name, filter by tag
- **Upload New** (within the picker) → inline upload modal (name + tags), assigns directly to the component without leaving the page
- **Edit Crops** → opens the crop editor for the active image
- Click the image count link to see all images assigned to a component

### Images tab

- Full library of uploaded images with name, filename, dimensions, and tags
- **Search** (300ms debounce) matches image names (strong) and filenames (weaker)
- **Tag chips** filter the grid — click to toggle, multiple tags use OR logic
- **Upload Image** → upload with a custom name and tags
- **Tags** → tag manager to create and delete tags
- **Delete** removes the image, its source file, Glide cache, and all generated output folders

## Image naming and tags

Every image has a **display name** (editable at upload, separate from the original filename) and one or more **tags**.

Tags are managed from the Images tab → **Tags** button. They can also be created inline during upload.

Example tags: `People`, `Outdoor`, `Product`, `Hero`

## Defining components and presets

Edit `config/components.json` and `config/presets.json` directly, or use the **Manage Components** / **Manage Presets** buttons in the topbar.

Components reference preset IDs:
```json
{
  "components": [
    { "id": "partner-banner", "label": "Partner Banner", "presets": ["homepage-banner"] }
  ]
}
```

Presets define output dimensions per variant:
```json
{
  "presets": [
    {
      "id": "homepage-banner",
      "label": "Homepage Banner",
      "format": "webp",
      "variants": [
        { "id": "mobile", "label": "Mobile", "width": 375, "height": 200 },
        { "id": "large", "label": "Large", "width": 1440, "height": 600 }
      ]
    }
  ]
}
```

## Output files

Generated files are organised by component and image name:

```
{output_folder}/
└── {component_id}/
    └── {image-name-slug}/
        ├── {preset_id}_{variant_id}_1x.webp
        └── {preset_id}_{variant_id}_2x.webp
```

Example for component `partner-banner`, image named "Summer Campaign", preset `homepage-banner`:

```
output/
└── partner-banner/
    └── summer-campaign/
        ├── homepage-banner_mobile_1x.webp
        ├── homepage-banner_mobile_2x.webp
        ├── homepage-banner_large_1x.webp
        └── homepage-banner_large_2x.webp
```

The **Open Folder** button in the crop editor opens exactly this subfolder in the OS file browser.

## Database schema

Three core tables + tag tables, all auto-migrated on first request:

| Table | Purpose |
|-------|---------|
| `images` | Uploaded image metadata (id, name, filename, dimensions, …) |
| `crops` | Saved crop coordinates per image × preset × variant |
| `component_images` | Many-to-many image ↔ component assignments |
| `tags` | Tag definitions (id, name) |
| `image_tags` | Many-to-many image ↔ tag pivot |

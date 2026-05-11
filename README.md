# ImageManager

Component-based image management and generation tool. Upload images once, crop interactively per preset, generate 1x/2x output variants automatically.

## What it does

1. **Upload** images → stored in `storage/source/`, metadata in SQLite
2. **Organize** images into *Components* (e.g. `partner-banner`, `product-hero`)
3. **Define** output *Presets* with variants (e.g. `homepage-banner` → mobile 375×200, large 1440×600)
4. **Crop** each image per variant using an interactive editor (Cropper.js)
5. **Generate** 1x and 2x output files → written to your configured output folder

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

On first request, the app auto-creates all required directories and sets the output folder to `{project}/output/`. No manual setup needed.

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
│   ├── index.php               # Component library view
│   ├── editor.php              # Crop editor
│   └── api/                   # REST API endpoints
├── src/                        # PHP application classes (App\ namespace)
│   ├── bootstrap.php           # Auto-setup (runs on every request via Composer)
│   ├── Database.php            # SQLite singleton + auto-migration
│   ├── GlideServer.php         # Glide image server factory
│   └── ...
├── storage/
│   ├── source/                 # Uploaded originals (gitignored)
│   └── cache/                  # Glide transform cache (gitignored)
└── vendor/
```

## Defining components and presets

Edit `config/components.json` and `config/presets.json` directly, or use the **Components** / **Presets** modals in the UI.

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

Generated files follow this naming convention:

```
{component_id}_{variant_id}_1x.webp
{component_id}_{variant_id}_2x.webp
```

Example: `partner-banner_mobile_1x.webp`, `partner-banner_mobile_2x.webp`

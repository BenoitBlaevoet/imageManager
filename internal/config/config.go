package config

import (
	"encoding/json"
	"fmt"
	"imagemanager/internal/model"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ── Presets ───────────────────────────────────────────────────────────────────

type presetsFile struct {
	Presets []model.Preset `json:"presets"`
}

func LoadPresets(appRoot string) ([]model.Preset, error) {
	data, err := os.ReadFile(filepath.Join(appRoot, "config", "presets.json"))
	if err != nil {
		return nil, err
	}
	var f presetsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.Presets, nil
}

func SavePresets(appRoot string, presets []model.Preset) error {
	f := presetsFile{Presets: presets}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appRoot, "config", "presets.json"), data, 0644)
}

func FindPreset(presets []model.Preset, id string) *model.Preset {
	for i := range presets {
		if presets[i].ID == id {
			return &presets[i]
		}
	}
	return nil
}

func FindVariant(preset *model.Preset, variantID string) *model.Variant {
	for i := range preset.Variants {
		if preset.Variants[i].ID == variantID {
			v := preset.Variants[i]
			if v.Format == "" {
				v.Format = preset.Format
			}
			return &v
		}
	}
	return nil
}

// ── Components ────────────────────────────────────────────────────────────────

type componentsFile struct {
	Components []model.Component `json:"components"`
}

func LoadComponents(appRoot string) ([]model.Component, error) {
	data, err := os.ReadFile(filepath.Join(appRoot, "config", "components.json"))
	if err != nil {
		return nil, err
	}
	var f componentsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.Components, nil
}

func SaveComponents(appRoot string, components []model.Component) error {
	f := componentsFile{Components: components}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appRoot, "config", "components.json"), data, 0644)
}

func FindComponent(components []model.Component, id string) *model.Component {
	for i := range components {
		if components[i].ID == id {
			return &components[i]
		}
	}
	return nil
}

// ── App config ────────────────────────────────────────────────────────────────

func LoadAppConfig(appRoot string) (model.AppConfig, error) {
	data, err := os.ReadFile(filepath.Join(appRoot, "config", "app-config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return model.AppConfig{DefaultFormat: "webp"}, nil
		}
		return model.AppConfig{}, err
	}
	var cfg model.AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return model.AppConfig{}, err
	}
	if cfg.DefaultFormat == "" {
		cfg.DefaultFormat = "webp"
	}
	return cfg, nil
}

func SaveAppConfig(appRoot string, cfg model.AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appRoot, "config", "app-config.json"), data, 0644)
}

// ── Validation ────────────────────────────────────────────────────────────────

var idRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func ValidatePresets(presets []model.Preset) []string {
	var errors []string
	seen := map[string]bool{}

	for i, p := range presets {
		prefix := fmt.Sprintf("presets[%d]", i)

		if p.ID == "" {
			errors = append(errors, prefix+".id is required and must be a non-empty string")
		} else if !idRegex.MatchString(p.ID) {
			errors = append(errors, prefix+".id must contain only lowercase letters, numbers, and hyphens")
		} else if seen[p.ID] {
			errors = append(errors, fmt.Sprintf("%s.id '%s' is already used", prefix, p.ID))
		} else {
			seen[p.ID] = true
		}

		if p.Label == "" {
			errors = append(errors, prefix+".label is required and must be a non-empty string")
		}

		if p.Format != "" && p.Format != "webp" && p.Format != "jpg" && p.Format != "png" {
			errors = append(errors, prefix+".format must be 'webp', 'jpg', or 'png'")
		}

		if len(p.Variants) == 0 {
			errors = append(errors, prefix+" must have at least one variant")
		} else {
			seenV := map[string]bool{}
			for j, v := range p.Variants {
				vprefix := fmt.Sprintf("%s.variants[%d]", prefix, j)
				if v.ID == "" {
					errors = append(errors, vprefix+".id is required and must be a non-empty string")
				} else if !idRegex.MatchString(v.ID) {
					errors = append(errors, vprefix+".id must contain only lowercase letters, numbers, and hyphens")
				} else if seenV[v.ID] {
					errors = append(errors, fmt.Sprintf("%s.id '%s' is already used in this preset", vprefix, v.ID))
				} else {
					seenV[v.ID] = true
				}
				if v.Label == "" {
					errors = append(errors, vprefix+".label is required and must be a non-empty string")
				}
				if v.Width <= 0 {
					errors = append(errors, vprefix+".width must be a positive integer")
				}
				if v.Height <= 0 {
					errors = append(errors, vprefix+".height must be a positive integer")
				}
				if v.Format != "" && v.Format != "webp" && v.Format != "jpg" && v.Format != "png" {
					errors = append(errors, vprefix+".format must be 'webp', 'jpg', or 'png'")
				}
			}
		}
	}
	return errors
}

func ValidateComponents(components []model.Component, presets []model.Preset) []string {
	var errors []string
	seen := map[string]bool{}

	for i, c := range components {
		prefix := fmt.Sprintf("components[%d]", i)

		if c.ID == "" {
			errors = append(errors, prefix+".id is required")
		} else if !idRegex.MatchString(c.ID) {
			errors = append(errors, prefix+".id must contain only lowercase letters, numbers, and hyphens")
		} else if seen[c.ID] {
			errors = append(errors, fmt.Sprintf("%s.id '%s' is already used", prefix, c.ID))
		} else {
			seen[c.ID] = true
		}

		if c.Label == "" {
			errors = append(errors, prefix+".label is required")
		}

		if len(c.Presets) == 0 {
			errors = append(errors, prefix+" must have at least one preset")
		} else {
			for _, pid := range c.Presets {
				if FindPreset(presets, pid) == nil {
					errors = append(errors, fmt.Sprintf("%s preset '%s' does not exist in presets.json", prefix, pid))
				}
			}
		}
	}
	return errors
}

// ── Helpers ───────────────────────────────────────────────────────────────────

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "image"
	}
	return s
}

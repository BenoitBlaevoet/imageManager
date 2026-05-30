package model

type Image struct {
	ID           string  `json:"id"`
	OriginalName string  `json:"original_name"`
	Name         string  `json:"name"`
	Filename     string  `json:"filename"`
	MimeType     string  `json:"mime_type"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	FileSize     int64   `json:"file_size"`
	CreatedAt    string  `json:"created_at"`
	GeneratedCount int   `json:"generated_count,omitempty"`
	CropCount    int     `json:"crop_count,omitempty"`
	Tags         []Tag   `json:"tags,omitempty"`
	ThumbURL     string  `json:"thumb_url,omitempty"`
}

type Crop struct {
	ID          int64   `json:"id"`
	ImageID     string  `json:"image_id"`
	PresetID    string  `json:"preset_id"`
	VariantID   string  `json:"variant_id"`
	CropX       int     `json:"crop_x"`
	CropY       int     `json:"crop_y"`
	CropW       int     `json:"crop_w"`
	CropH       int     `json:"crop_h"`
	GeneratedAt *string `json:"generated_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type Tag struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	ImageCount int    `json:"image_count,omitempty"`
}

type ComponentImage struct {
	AssignmentID int64  `json:"assignment_id"`
	ComponentID  string `json:"component_id"`
	ImageID      string `json:"id"`
	IsActive     bool   `json:"is_active"`
	AssignedAt   string `json:"assigned_at"`
	OriginalName string `json:"original_name"`
	Name         string `json:"name"`
	Filename     string `json:"filename"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int64  `json:"file_size"`
	CreatedAt    string `json:"created_at"`
	Tags         []Tag  `json:"tags,omitempty"`
	ThumbURL     string `json:"thumb_url,omitempty"`
}

type Preset struct {
	ID       string    `json:"id"`
	Label    string    `json:"label"`
	Format   string    `json:"format,omitempty"`
	Variants []Variant `json:"variants"`
}

type Variant struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format,omitempty"`
}

type Component struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Presets     []string `json:"presets"`
}

type AppConfig struct {
	OutputFolder  string `json:"output_folder"`
	DefaultFormat string `json:"default_format"`
}

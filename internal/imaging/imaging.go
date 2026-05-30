package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"

	genwebp "github.com/gen2brain/webp"
	"github.com/disintegration/imaging"
)

// DecodeAny decodes an image.Image from a reader given the MIME type.
// WebP files use gen2brain/webp (pure Go via WASM); all others use imaging.Decode.
func DecodeAny(r io.Reader, mimeType string) (image.Image, error) {
	if mimeType == "image/webp" {
		return genwebp.Decode(r)
	}
	return imaging.Decode(r)
}

// CropAndResize crops a region from src (source pixel coords) then resizes to targetW×targetH.
func CropAndResize(src image.Image, cropX, cropY, cropW, cropH, targetW, targetH int) image.Image {
	cropped := imaging.Crop(src, image.Rect(cropX, cropY, cropX+cropW, cropY+cropH))
	return imaging.Resize(cropped, targetW, targetH, imaging.Lanczos)
}

// ThumbnailFit resizes src to fit within w×h preserving aspect ratio.
// Pass h=0 to scale by width only.
func ThumbnailFit(src image.Image, w, h int) image.Image {
	if h == 0 {
		return imaging.Resize(src, w, 0, imaging.Lanczos)
	}
	return imaging.Fit(src, w, h, imaging.Lanczos)
}

// ThumbnailFill resizes src to exactly w×h via center crop (no distortion).
func ThumbnailFill(src image.Image, w, h int) image.Image {
	return imaging.Fill(src, w, h, imaging.Center, imaging.Linear)
}

// EncodeImage encodes img to w in the requested format (webp/png/jpg).
func EncodeImage(out io.Writer, img image.Image, format string) error {
	switch format {
	case "webp":
		return genwebp.Encode(out, img, genwebp.Options{Quality: 85})
	case "png":
		return png.Encode(out, img)
	default: // jpg / jpeg
		return jpeg.Encode(out, compositeOnWhite(img), &jpeg.Options{Quality: 85})
	}
}

// OutputExt returns the file extension for the given format string.
func OutputExt(format string) string {
	switch format {
	case "webp":
		return "webp"
	case "png":
		return "png"
	default:
		return "jpg"
	}
}

// EncodeJPEG encodes img as JPEG (used for thumbnails/display served over HTTP).
func EncodeJPEG(out io.Writer, img image.Image) error {
	return jpeg.Encode(out, compositeOnWhite(img), &jpeg.Options{Quality: 85})
}

// EncodeJPEGBytes encodes img as JPEG and returns the raw bytes.
func EncodeJPEGBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := EncodeJPEG(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// compositeOnWhite flattens the alpha channel onto a white background.
// Required before JPEG encoding to avoid dark transparent regions.
func compositeOnWhite(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
	return dst
}

package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp" // register WebP decode
)

// DecodeAny decodes an image.Image from a reader, given the MIME type.
func DecodeAny(r io.Reader, _ string) (image.Image, error) {
	// imaging.Decode handles jpeg, png, gif, tiff, bmp.
	// WebP is handled by golang.org/x/image/webp registered via the blank import.
	return imaging.Decode(r)
}

// CropAndResize crops a region from src (in source coordinates) then resizes to targetW×targetH.
func CropAndResize(src image.Image, cropX, cropY, cropW, cropH, targetW, targetH int) image.Image {
	cropped := imaging.Crop(src, image.Rect(cropX, cropY, cropX+cropW, cropY+cropH))
	return imaging.Resize(cropped, targetW, targetH, imaging.Lanczos)
}

// ThumbnailFit resizes src to fit within w×h, preserving aspect ratio.
// If h is 0, scales by width only.
func ThumbnailFit(src image.Image, w, h int) image.Image {
	if h == 0 {
		return imaging.Resize(src, w, 0, imaging.Lanczos)
	}
	return imaging.Fit(src, w, h, imaging.Lanczos)
}

// ThumbnailFill resizes src to exactly w×h using center crop (no distortion).
func ThumbnailFill(src image.Image, w, h int) image.Image {
	return imaging.Fill(src, w, h, imaging.Center, imaging.Linear)
}

// EncodeImage encodes img to the writer in the requested format.
// WebP output is not supported without CGo; webp requests fall back to JPEG.
func EncodeImage(w io.Writer, img image.Image, format string) error {
	switch format {
	case "png":
		return png.Encode(w, img)
	default: // jpg / jpeg / webp (webp falls back to jpeg)
		return jpeg.Encode(w, compositeOnWhite(img), &jpeg.Options{Quality: 85})
	}
}

// OutputExt returns the file extension for the given format.
// webp falls back to jpg since we can't encode WebP without CGo.
func OutputExt(format string) string {
	switch format {
	case "png":
		return "png"
	default:
		return "jpg"
	}
}

// EncodeJPEG encodes img as JPEG (for thumbnails/display).
func EncodeJPEG(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, compositeOnWhite(img), &jpeg.Options{Quality: 85})
}

// EncodeJPEGBytes encodes img as JPEG and returns the bytes.
func EncodeJPEGBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := EncodeJPEG(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// compositeOnWhite flattens alpha onto white (required before JPEG encoding).
func compositeOnWhite(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
	return dst
}

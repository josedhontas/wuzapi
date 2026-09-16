package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestLinkPreviewThumbnailContainsPortraitImage(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 100, 200))
	fillImage(source, color.RGBA{R: 220, G: 20, B: 20, A: 255})

	thumbnail := createLinkPreviewThumbnail(t, source)

	assertMostlyWhite(t, thumbnail.At(100, manualLinkPreviewThumbnailHeight/2))
	assertMostlyRed(t, thumbnail.At(manualLinkPreviewThumbnailWidth/2, 5))
	assertMostlyRed(t, thumbnail.At(manualLinkPreviewThumbnailWidth/2, manualLinkPreviewThumbnailHeight-6))
}

func TestLinkPreviewThumbnailContainsWideImage(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 400, 100))
	fillImage(source, color.RGBA{R: 20, G: 160, B: 30, A: 255})

	thumbnail := createLinkPreviewThumbnail(t, source)

	assertMostlyWhite(t, thumbnail.At(manualLinkPreviewThumbnailWidth/2, 20))
	assertMostlyGreen(t, thumbnail.At(5, manualLinkPreviewThumbnailHeight/2))
	assertMostlyGreen(t, thumbnail.At(manualLinkPreviewThumbnailWidth-6, manualLinkPreviewThumbnailHeight/2))
}

func createLinkPreviewThumbnail(t *testing.T, source image.Image) image.Image {
	t.Helper()

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, source); err != nil {
		t.Fatalf("encode source image: %v", err)
	}

	data, width, height, err := linkPreviewThumbnailFromBytes(encoded.Bytes())
	if err != nil {
		t.Fatalf("create link preview thumbnail: %v", err)
	}
	if width != manualLinkPreviewThumbnailWidth || height != manualLinkPreviewThumbnailHeight {
		t.Fatalf("unexpected thumbnail dimensions: %dx%d", width, height)
	}

	thumbnail, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode thumbnail: %v", err)
	}
	return thumbnail
}

func fillImage(img *image.RGBA, fill color.Color) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.Set(x, y, fill)
		}
	}
}

func assertMostlyWhite(t *testing.T, pixel color.Color) {
	t.Helper()
	r, g, b, _ := pixel.RGBA()
	if r < 0xd000 || g < 0xd000 || b < 0xd000 {
		t.Fatalf("expected a white margin, got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func assertMostlyRed(t *testing.T, pixel color.Color) {
	t.Helper()
	r, g, b, _ := pixel.RGBA()
	if r < 0x9000 || g > 0x6000 || b > 0x6000 {
		t.Fatalf("expected image content to remain red, got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func assertMostlyGreen(t *testing.T, pixel color.Color) {
	t.Helper()
	r, g, b, _ := pixel.RGBA()
	if g < 0x7000 || r > 0x6000 || b > 0x6000 {
		t.Fatalf("expected image content to remain green, got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

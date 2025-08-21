package core

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

// MockJPEGImageProcessor implements ImageProcessor for testing
type MockJPEGImageProcessor struct{}

func NewMockJPEGImageProcessor() *MockJPEGImageProcessor {
	return &MockJPEGImageProcessor{}
}

// GetDimensions returns mock dimensions for testing
func (p *MockJPEGImageProcessor) GetDimensions(imagePath string) (*ImageDimensions, error) {
	// Return large test dimensions to avoid capacity constraints
	return &ImageDimensions{
		Width:  1024,
		Height: 1024,
	}, nil
}

// DecodeJPEG creates a mock JPEG image for testing
func (p *MockJPEGImageProcessor) DecodeJPEG(imagePath string) (image.Image, error) {
	// Create a mock 512x512 image
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))

	// Fill with a gradient pattern
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x + y) % 256),
				G: uint8(x % 256),
				B: uint8(y % 256),
				A: 255,
			})
		}
	}

	return img, nil
}

// EncodePNG saves the image as PNG
func (p *MockJPEGImageProcessor) EncodePNG(img image.Image, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Encode as JPEG for simplicity
	return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
}

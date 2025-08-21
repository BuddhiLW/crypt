package core

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
)

// JPEGImageProcessor implements ImageProcessor for JPEG images
type JPEGImageProcessor struct{}

func NewJPEGImageProcessor() *JPEGImageProcessor {
	return &JPEGImageProcessor{}
}

func (p *JPEGImageProcessor) GetDimensions(imagePath string) (*ImageDimensions, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	config, err := jpeg.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	return &ImageDimensions{
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

func (p *JPEGImageProcessor) DecodeJPEG(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JPEG: %w", err)
	}

	return img, nil
}

func (p *JPEGImageProcessor) EncodePNG(img image.Image, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create PNG file: %w", err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

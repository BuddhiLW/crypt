package core

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/liyue201/goqr"
	"github.com/skip2/go-qrcode"
)

// GoQRProcessor implements QRCodeProcessor using go-qrcode and goqr libraries
type GoQRProcessor struct{}

func NewGoQRProcessor() *GoQRProcessor {
	return &GoQRProcessor{}
}

func (p *GoQRProcessor) GenerateQR(data string, size int, eccLevel ECCLevel) ([]byte, error) {
	// Convert our ECC level to go-qrcode level
	var qrECC qrcode.RecoveryLevel
	switch eccLevel {
	case ECCLevelLow:
		qrECC = qrcode.Low
	case ECCLevelMedium:
		qrECC = qrcode.Medium
	case ECCLevelHigh:
		qrECC = qrcode.High
	case ECCLevelHighest:
		qrECC = qrcode.Highest
	default:
		qrECC = qrcode.High // Default to High for robustness
	}

	// Generate QR code
	qr, err := qrcode.New(data, qrECC)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Generate PNG bytes
	pngBytes, err := qr.PNG(size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PNG: %w", err)
	}

	return pngBytes, nil
}

func (p *GoQRProcessor) ReadQR(imagePath string) (string, error) {
	// Read image file
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Recognize QR codes
	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		return "", fmt.Errorf("failed to recognize QR code: %w", err)
	}

	if len(qrCodes) == 0 {
		return "", errors.New("no QR code found in image")
	}

	return string(qrCodes[0].Payload), nil
}

func (p *GoQRProcessor) ConvertToBitstream(pngData []byte) ([]byte, error) {
	// Decode PNG
	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, errors.New("failed to decode PNG image")
	}

	// Convert to grayscale
	grayImg := image.NewGray(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}

	// Get dimensions
	bounds := grayImg.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Prepare bitstream
	bitstream := make([]byte, (width*height+7)/8)
	bitIndex := 0

	// Convert pixels to bits
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := grayImg.GrayAt(x, y).Y

			// Black pixels are '1', white pixels are '0'
			if pixel < 128 {
				bitstream[bitIndex/8] |= 1 << (7 - (bitIndex % 8))
			}
			bitIndex++
		}
	}

	return bitstream, nil
}

func (p *GoQRProcessor) ConvertFromBitstream(bitstream []byte, size int) (image.Image, error) {
	img := image.NewGray(image.Rect(0, 0, size, size))
	bitIndex := 0

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if bitIndex < len(bitstream)*8 {
				bit := (bitstream[bitIndex/8] >> (7 - (bitIndex % 8))) & 1
				if bit == 1 {
					img.SetGray(x, y, color.Gray{Y: 0}) // Black
				} else {
					img.SetGray(x, y, color.Gray{Y: 255}) // White
				}
			}
			bitIndex++
		}
	}

	return img, nil
}

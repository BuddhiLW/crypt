package core

import (
	"fmt"
	"math"
)

// StandardQRSizeCalculator implements QRSizeCalculator with standard QR capacity mapping
type StandardQRSizeCalculator struct {
	imageProcessor ImageProcessor
}

func NewStandardQRSizeCalculator() *StandardQRSizeCalculator {
	return &StandardQRSizeCalculator{
		imageProcessor: NewJPEGImageProcessor(),
	}
}

func NewStandardQRSizeCalculatorWithProcessor(processor ImageProcessor) *StandardQRSizeCalculator {
	return &StandardQRSizeCalculator{
		imageProcessor: processor,
	}
}

func (c *StandardQRSizeCalculator) CalculateOptimalSize(imagePath string, payloadSize int, strategy DCTStrategy) (int, error) {
	// Get image dimensions
	dims, err := c.imageProcessor.GetDimensions(imagePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get image dimensions: %w", err)
	}

	// Calculate DCT capacity using the strategy
	dctCapacityBits := c.calculateDCTCapacity(dims.Width, dims.Height, strategy)

	// Calculate maximum QR size that fits in DCT capacity
	maxQRPixelsFromDCT := int(float64(dctCapacityBits) * 0.9) // Use 90% of capacity for safety
	maxQRSizeFromDCT := int(math.Sqrt(float64(maxQRPixelsFromDCT)))
	maxQRSizeFromDCT = (maxQRSizeFromDCT / 8) * 8 // Round down to multiple of 8

	// Calculate 80% of the smaller image dimension
	minDim := dims.Width
	if dims.Height < minDim {
		minDim = dims.Height
	}
	qrSizeFromDim := int(float64(minDim) * 0.8)

	// Calculate optimal QR size for High/Highest ECC only
	optimalQRSize, err := c.findOptimalQRSizeForHighECC(payloadSize)
	if err != nil {
		return 0, fmt.Errorf("payload too large for High ECC: %w", err)
	}

	// Use the optimal size, but check against capacity limits
	qrSize := optimalQRSize
	if qrSize > qrSizeFromDim {
		qrSize = qrSizeFromDim
	}
	if qrSize > maxQRSizeFromDCT {
		qrSize = maxQRSizeFromDCT
	}

	// Verify the final size can still hold the payload with High ECC
	if qrSize < optimalQRSize {
		return 0, fmt.Errorf("image constraints prevent using optimal QR size %dx%d for High ECC (limited to %dx%d)",
			optimalQRSize, optimalQRSize, qrSize, qrSize)
	}

	// Round to nearest multiple of 8 for better alignment
	qrSize = (qrSize / 8) * 8

	// Ensure minimum size of 64 and maximum of 512
	if qrSize < 64 {
		qrSize = 64
	}
	if qrSize > 512 {
		qrSize = 512
	}

	return qrSize, nil
}

func (c *StandardQRSizeCalculator) GetCapacityMap() map[int][2]int {
	return map[int][2]int{
		64:  {15, 30},     // Highest: ~15 bytes, High: ~30 bytes
		96:  {40, 80},     // Highest: ~40 bytes, High: ~80 bytes
		128: {75, 150},    // Highest: ~75 bytes, High: ~150 bytes
		160: {125, 250},   // Highest: ~125 bytes, High: ~250 bytes
		192: {175, 350},   // Highest: ~175 bytes, High: ~350 bytes
		224: {250, 500},   // Highest: ~250 bytes, High: ~500 bytes
		256: {350, 700},   // Highest: ~350 bytes, High: ~700 bytes
		288: {450, 900},   // Highest: ~450 bytes, High: ~900 bytes
		320: {600, 1200},  // Highest: ~600 bytes, High: ~1.2KB
		352: {750, 1500},  // Highest: ~750 bytes, High: ~1.5KB
		384: {900, 1800},  // Highest: ~900 bytes, High: ~1.8KB
		416: {1100, 2200}, // Highest: ~1.1KB, High: ~2.2KB
		448: {1300, 2600}, // Highest: ~1.3KB, High: ~2.6KB
		480: {1500, 3000}, // Highest: ~1.5KB, High: ~3KB
		512: {1750, 3500}, // Highest: ~1.75KB, High: ~3.5KB
	}
}

func (c *StandardQRSizeCalculator) calculateDCTCapacity(width, height int, strategy DCTStrategy) int {
	dctBlocksX := (width + 7) / 8
	dctBlocksY := (height + 7) / 8
	return dctBlocksX * dctBlocksY * strategy.GetCoefficientsPerBit()
}

func (c *StandardQRSizeCalculator) findOptimalQRSizeForHighECC(payloadBytes int) (int, error) {
	capacityMap := c.GetCapacityMap()

	// Try to find a QR size that can hold the payload with Highest ECC first
	for size := 64; size <= 512; size += 32 {
		if capacities, exists := capacityMap[size]; exists {
			// Try Highest ECC first
			if capacities[0] >= payloadBytes {
				return size, nil
			}
			// Try High ECC as fallback
			if capacities[1] >= payloadBytes {
				return size, nil
			}
		}
	}

	return 0, fmt.Errorf("payload %d bytes exceeds High ECC capacity at maximum QR size (512x512, ~3.5KB max)", payloadBytes)
}

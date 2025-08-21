package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rwxrob/bonzai/vars"
)

// SteganographyService orchestrates the complete steganography workflow
type SteganographyService struct {
	imageProcessor  ImageProcessor
	qrProcessor     QRCodeProcessor
	dctProcessor    DCTProcessor
	sizeCalculator  QRSizeCalculator
	metadataManager MetadataManager
}

// NewSteganographyService creates a new service with all dependencies
func NewSteganographyService(
	imageProcessor ImageProcessor,
	qrProcessor QRCodeProcessor,
	dctProcessor DCTProcessor,
	sizeCalculator QRSizeCalculator,
	metadataManager MetadataManager,
) *SteganographyService {
	return &SteganographyService{
		imageProcessor:  imageProcessor,
		qrProcessor:     qrProcessor,
		dctProcessor:    dctProcessor,
		sizeCalculator:  sizeCalculator,
		metadataManager: metadataManager,
	}
}

// EmbedQRCode embeds a QR code into a JPEG image
func (s *SteganographyService) EmbedQRCode(inputPath, outputPath, data string, strategy DCTStrategy, env string) error {
	// Try to get QR size from vars first, fall back to calculation
	qrSizeStr, _ := vars.Get("qr-size", "DCT_ENV")
	var qrSize int
	var err error

	if qrSizeStr != "" {
		qrSize, err = strconv.Atoi(qrSizeStr)
		if err != nil {
			fmt.Printf("WARNING: Invalid QR size in vars '%s', falling back to calculation: %v\n", qrSizeStr, err)
			qrSize, err = s.sizeCalculator.CalculateOptimalSize(inputPath, len(data), strategy)
			if err != nil {
				return fmt.Errorf("failed to calculate QR size: %w", err)
			}
		} else {
			fmt.Printf("DEBUG: Using QR size from vars: %d\n", qrSize)
		}
	} else {
		// Calculate optimal QR size
		qrSize, err = s.sizeCalculator.CalculateOptimalSize(inputPath, len(data), strategy)
		if err != nil {
			return fmt.Errorf("failed to calculate QR size: %w", err)
		}
		fmt.Printf("DEBUG: Using calculated QR size: %d\n", qrSize)
	}

	// Generate QR code with High ECC
	qrPNG, err := s.qrProcessor.GenerateQR(data, qrSize, ECCLevelHigh)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Debug: Save the generated QR code to a file for inspection
	debugQRPath := filepath.Join(filepath.Dir(outputPath), "debug_generated_qr.png")
	if err := os.WriteFile(debugQRPath, qrPNG, 0644); err != nil {
		fmt.Printf("WARNING: Failed to save debug QR code: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Saved generated QR code to %s\n", debugQRPath)
	}

	// Convert PNG to bitstream
	bitstream, err := s.qrProcessor.ConvertToBitstream(qrPNG)
	if err != nil {
		return fmt.Errorf("failed to convert QR to bitstream: %w", err)
	}

	// Get actual dimensions from generated QR
	img, err := s.imageProcessor.DecodeJPEG(inputPath)
	if err != nil {
		return fmt.Errorf("failed to decode input image: %w", err)
	}
	actualQRSize := img.Bounds().Dx()

	// Calculate actual data area (83% of pixels contain actual data)
	actualDataArea := int(float64(actualQRSize) * 0.83)
	actualDataArea = (actualDataArea / 8) * 8 // Round to multiple of 8

	// Store metadata
	err = s.metadataManager.StoreQRMetadata(actualQRSize, len(bitstream), actualDataArea, env)
	if err != nil {
		return fmt.Errorf("failed to store QR metadata: %w", err)
	}

	err = s.metadataManager.StoreDCTStrategy(strategy, env)
	if err != nil {
		return fmt.Errorf("failed to store DCT strategy: %w", err)
	}

	// Embed bitstream using DCT
	err = s.dctProcessor.EmbedData(inputPath, outputPath, bitstream, strategy)
	if err != nil {
		return fmt.Errorf("failed to embed data: %w", err)
	}

	return nil
}

// ExtractQRCode extracts a QR code from a JPEG image
func (s *SteganographyService) ExtractQRCode(inputPath, outputPath, env string) error {
	// Retrieve metadata
	pixelSize, _, _, err := s.metadataManager.RetrieveQRMetadata(env)
	if err != nil {
		return fmt.Errorf("failed to retrieve QR metadata: %w", err)
	}

	strategy, err := s.metadataManager.RetrieveDCTStrategy(env)
	if err != nil {
		return fmt.Errorf("failed to retrieve DCT strategy: %w", err)
	}

	// Calculate buffer size for extraction
	fullPixelSize := pixelSize * pixelSize / 8
	qrBitstream := make([]byte, fullPixelSize)

	// Extract bitstream using DCT
	extractedData, err := s.dctProcessor.ExtractData(inputPath, fullPixelSize, strategy)
	if err != nil {
		return fmt.Errorf("failed to extract data: %w", err)
	}

	// Copy extracted data to our buffer
	copy(qrBitstream, extractedData)

	// Convert bitstream back to QR image
	img, err := s.qrProcessor.ConvertFromBitstream(qrBitstream, pixelSize)
	if err != nil {
		return fmt.Errorf("failed to convert bitstream to QR image: %w", err)
	}

	// Save QR code as PNG
	err = s.imageProcessor.EncodePNG(img, outputPath)
	if err != nil {
		return fmt.Errorf("failed to save QR image: %w", err)
	}

	return nil
}

// ReadQRCode reads a QR code from an image file
func (s *SteganographyService) ReadQRCode(imagePath string) (string, error) {
	return s.qrProcessor.ReadQR(imagePath)
}

// EmbedMultiQRWithMetadata embeds data as multiple QR codes with hash-based metadata
func (s *SteganographyService) EmbedMultiQRWithMetadata(inputPath, outputDir, data, env string) error {
	fmt.Printf("DEBUG: ===== EmbedMultiQRWithMetadata START =====\n")
	fmt.Printf("DEBUG: EmbedMultiQRWithMetadata called with inputPath=%s, outputDir=%s, dataLength=%d\n",
		inputPath, outputDir, len(data))

	// Create temporary directory for multi-QR files following Bonzai patterns
	tempDir, err := os.MkdirTemp("", "crypt-multiqr-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("WARNING: Failed to clean up temp directory %s: %v\n", tempDir, err)
		}
	}()

	fmt.Printf("DEBUG: Created temp directory: %s\n", tempDir)

	// Import the encrypt package function for embedding QR codes
	// We'll use a direct approach to avoid circular dependencies

	// Create metadata file in temp directory
	metadataPath := filepath.Join(tempDir, "metadata.jpeg")
	fmt.Printf("DEBUG: Creating metadata file at %s\n", metadataPath)

	// Calculate optimal chunk size and count first
	// QR code can store ~50-80 characters in 96x96 QR code with high ECC
	// Use 50 bytes per chunk for safety with 96x96 QR codes
	maxChunkSize := 50
	dataLen := len(data)
	chunkCount := (dataLen + maxChunkSize - 1) / maxChunkSize // Ceiling division

	fmt.Printf("DEBUG: Data length: %d bytes, splitting into %d chunks of max %d bytes each\n", dataLen, chunkCount, maxChunkSize)

	// Create proper metadata structure for multi-QR system
	metadata := map[string]interface{}{
		"grid_width":  1,
		"grid_height": 1,
		"chunk_count": chunkCount,
		"chunk_size":  maxChunkSize,
		"total_size":  dataLen,
		"qr_size":     96,
		"padding":     24,
	}

	metadataJSON, jsonErr := json.Marshal(metadata)
	if jsonErr != nil {
		return fmt.Errorf("failed to create metadata JSON: %w", jsonErr)
	}

	// Store QR size in vars for metadata embedding (use the same env as extraction)
	if err := vars.Set("qr-size", "96", "DCT_ENV"); err != nil {
		fmt.Printf("WARNING: Failed to store QR size for metadata: %v\n", err)
	}
	if err := vars.Set("QR_DATA_AREA", "72", "DCT_ENV"); err != nil {
		fmt.Printf("WARNING: Failed to store QR data area for metadata: %v\n", err)
	}

	err = embedQRCodeInJPEG(inputPath, metadataPath, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	fmt.Printf("DEBUG: Successfully created metadata file\n")

	// Use the calculated chunk count and sizes

	// Create chunk files in temp directory
	for i := 0; i < chunkCount; i++ {
		chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d.jpeg", i))
		fmt.Printf("DEBUG: Creating chunk file %d at %s\n", i, chunkPath)

		// Calculate chunk boundaries
		start := i * maxChunkSize
		end := start + maxChunkSize
		if end > dataLen {
			end = dataLen
		}
		chunkData := string(data[start:end])

		fmt.Printf("DEBUG: Chunk %d: bytes %d-%d (length: %d), first 50 chars: %s\n",
			i, start, end-1, len(chunkData), chunkData[:min(50, len(chunkData))])

		// Store QR size in vars for extraction (use the same env as extraction)
		if err := vars.Set("qr-size", "96", "DCT_ENV"); err != nil {
			fmt.Printf("WARNING: Failed to store QR size: %v\n", err)
		}
		if err := vars.Set("QR_DATA_AREA", "72", "DCT_ENV"); err != nil {
			fmt.Printf("WARNING: Failed to store QR data area: %v\n", err)
		}

		err = embedQRCodeInJPEG(inputPath, chunkPath, chunkData)
		if err != nil {
			return fmt.Errorf("failed to create chunk file %d: %w", i, err)
		}
		fmt.Printf("DEBUG: Successfully created chunk file %d\n", i)
	}

	// Copy all files from temp directory to output directory
	if err := copyDirectory(tempDir, outputDir); err != nil {
		return fmt.Errorf("failed to copy files to output directory: %w", err)
	}

	fmt.Printf("DEBUG: Successfully copied all files to %s\n", outputDir)
	return nil
}

// embedQRCodeInJPEG embeds QR code data into a JPEG file
func embedQRCodeInJPEG(inputPath, outputPath, qrData string) error {
	fmt.Printf("DEBUG: embedQRCodeInJPEG called with inputPath=%s, outputPath=%s, qrDataLength=%d\n",
		inputPath, outputPath, len(qrData))

	// Use the factory to create a service with the real DCT processor
	factory := NewServiceFactory()
	service := factory.CreateSteganographyService("multiqr") // Use multiqr env

	fmt.Printf("DEBUG: Created service for QR embedding\n")

	err := service.EmbedQRCode(inputPath, outputPath, qrData, DCTStrategySingle, "multiqr")
	if err != nil {
		return fmt.Errorf("failed to embed QR code: %w", err)
	}

	fmt.Printf("DEBUG: Successfully embedded QR code\n")
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// copyDirectory copies all files from src to dst
func copyDirectory(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := dstFile.ReadFrom(srcFile); err != nil {
			return err
		}
	}

	return nil
}

// ExtractMultiQRWithMetadata extracts data from multiple QR codes using metadata
func (s *SteganographyService) ExtractMultiQRWithMetadata(metadataFile, chunkDir, key, env string) (string, error) {
	// This will be implemented to use the enhanced multi-QR functionality
	return "", fmt.Errorf("enhanced multi-QR extraction not yet implemented")
}

// ScanAndExtractMultiQR scans directory and automatically extracts data
func (s *SteganographyService) ScanAndExtractMultiQR(directory, key, env string) (string, error) {
	// This will be implemented to use the enhanced multi-QR functionality
	return "", fmt.Errorf("enhanced multi-QR scanning not yet implemented")
}

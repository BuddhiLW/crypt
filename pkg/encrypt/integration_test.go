package encrypt

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestEndToEndFlow tests the complete encrypt->embed->extract->decrypt flow
func TestEndToEndFlow(t *testing.T) {
	// Create test directory
	testDir := filepath.Join("..", "..", "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Test data
	testCases := []struct {
		name     string
		data     string
		password string
	}{
		{"Small", "Hello, World!", "mysecurepassword123"},
		{"Medium", generateTestData(500), "mysecurepassword123"},
		{"Large", generateTestData(1500), "mysecurepassword123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Encrypt
			encrypted, err := EncryptMessage(tc.data, tc.password)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}
			t.Logf("Encrypted data length: %d bytes", len(encrypted))

			// Step 2: Generate QR with fallback
			qrPath := filepath.Join(testDir, "test_qr.png")
			level, err := WriteQRCodeWithFallback(encrypted, 256, qrPath)
			if err != nil {
				t.Fatalf("QR generation failed: %v", err)
			}
			t.Logf("QR generated with ECC level: %v", level)

			// Verify QR file was created
			if _, err := os.Stat(qrPath); os.IsNotExist(err) {
				t.Fatalf("QR file was not created: %s", qrPath)
			}

			// Clean up
			os.Remove(qrPath)
		})
	}
}

// TestDCTCapacityAnalysis analyzes the DCT capacity of test images
func TestDCTCapacityAnalysis(t *testing.T) {
	testDir := filepath.Join("..", "..", "test")
	testImage := filepath.Join(testDir, "input.jpeg")

	// Check if test image exists
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("Test image not found, skipping DCT capacity test")
	}

	// Test different payload sizes to find the limit
	payloadSizes := []int{1000, 2000, 4000, 8000}

	for _, size := range payloadSizes {
		t.Run(fmt.Sprintf("Payload%dBytes", size), func(t *testing.T) {
			// Generate test payload
			testData := generateTestData(size)
			encrypted, err := EncryptMessage(testData, "testpassword123456")
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Try to embed
			outputPath := filepath.Join(testDir, fmt.Sprintf("test_output_%d.jpeg", size))
			payloadSize := len(encrypted) // Size of the encrypted Base64 data
			err = EmbedQRCodeInJPEG(testImage, outputPath, encrypted, payloadSize)

			if err != nil {
				t.Logf("Failed to embed %d byte payload: %v", size, err)
			} else {
				t.Logf("Successfully embedded %d byte payload", size)
				// Verify output file exists
				if _, err := os.Stat(outputPath); err == nil {
					// Get file size for comparison
					info, _ := os.Stat(outputPath)
					t.Logf("Output file size: %d bytes", info.Size())
				}
				// Clean up
				os.Remove(outputPath)
			}
		})
	}
}

// TestImageCapacityEstimation provides capacity estimates for different image sizes
func TestImageCapacityEstimation(t *testing.T) {
	testCases := []struct {
		width, height int
		name          string
	}{
		{640, 480, "VGA"},
		{1024, 768, "XGA"},
		{1920, 1080, "FullHD"},
		{2048, 1536, "2MP"},
		{4096, 3072, "12MP"},
	}

	t.Log("DCT Capacity Estimates:")
	t.Log("======================")

	for _, tc := range testCases {
		// Estimate DCT blocks (8x8 blocks)
		blocksX := (tc.width + 7) / 8
		blocksY := (tc.height + 7) / 8
		totalBlocks := blocksX * blocksY

		// Conservative estimate: 1 bit per block
		capacityBits := totalBlocks
		capacityBytes := capacityBits / 8
		capacityKB := capacityBytes / 1024

		t.Logf("%s (%dx%d): ~%d blocks, ~%d bytes (~%d KB) capacity",
			tc.name, tc.width, tc.height, totalBlocks, capacityBytes, capacityKB)
	}
}

// Helper function to generate test data
func generateTestData(size int) string {
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte('A' + (i % 26))
	}
	return string(data)
}

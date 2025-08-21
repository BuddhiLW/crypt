package decrypt_test

import (
	"os"
	"testing"
)

func TestDCTExtraction(t *testing.T) {
	// Test data setup
	testImagePath := "../../test/simple_qr_test.jpeg"
	tempQRPath := "/tmp/test_extracted_qr.png"

	// Ensure test image exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skipf("Test image not found: %s", testImagePath)
	}

	// Clean up temp file
	defer os.Remove(tempQRPath)

	t.Run("ExtractQRCodeFromJPEG", func(t *testing.T) {
		err := ExtractQRCodeFromJPEG(testImagePath, tempQRPath)
		if err != nil {
			t.Errorf("ExtractQRCodeFromJPEG failed: %v", err)
			return
		}

		// Check if output file was created
		if _, err := os.Stat(tempQRPath); os.IsNotExist(err) {
			t.Errorf("Output QR file was not created: %s", tempQRPath)
		}

		// Check file size
		info, err := os.Stat(tempQRPath)
		if err != nil {
			t.Errorf("Failed to stat output file: %v", err)
		} else if info.Size() == 0 {
			t.Errorf("Output QR file is empty")
		} else {
			t.Logf("Output QR file created: %s (%d bytes)", tempQRPath, info.Size())
		}
	})
}

func TestQRCodeReading(t *testing.T) {
	// Test the QR code reading function separately
	testQRPath := "/tmp/qr.png" // This should exist from previous runs

	if _, err := os.Stat(testQRPath); os.IsNotExist(err) {
		t.Skipf("Test QR image not found: %s", testQRPath)
	}

	t.Run("readQRCode", func(t *testing.T) {
		data, err := readQRCode(testQRPath)
		if err != nil {
			t.Errorf("readQRCode failed: %v", err)
			return
		}

		if len(data) == 0 {
			t.Errorf("readQRCode returned empty data")
		} else {
			t.Logf("readQRCode success: %d bytes", len(data))
			t.Logf("First 50 chars: %q", truncateString(data, 50))
		}
	})
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func TestConvertBitstreamToQRImage(t *testing.T) {
	// Test bitstream to image conversion
	t.Run("ConvertBitstreamToQRImage", func(t *testing.T) {
		// Create test bitstream (simple pattern)
		testBitstream := make([]byte, 128*128/8) // 128x128 QR code

		// Set some bits to create a pattern
		testBitstream[0] = 0xFF // First 8 bits set
		testBitstream[1] = 0xAA // Alternating pattern
		testBitstream[2] = 0x55 // Alternating pattern

		img, err := ConvertBitstreamToQRImage(testBitstream, 128)
		if err != nil {
			t.Errorf("ConvertBitstreamToQRImage failed: %v", err)
			return
		}

		bounds := img.Bounds()
		if bounds.Dx() != 128 || bounds.Dy() != 128 {
			t.Errorf("Image size mismatch: got %dx%d, want 128x128", bounds.Dx(), bounds.Dy())
		}

		t.Logf("ConvertBitstreamToQRImage success: %dx%d image", bounds.Dx(), bounds.Dy())
	})
}

package decrypt_test

import (
	"image"
	"testing"
)

func TestBitstreamAnalysis(t *testing.T) {
	testImagePath := "../../test/simple_qr_test.jpeg"

	t.Run("AnalyzeFullBitstream", func(t *testing.T) {
		// Extract the full bitstream for analysis
		tempQRPath := "/tmp/bitstream_analysis.png"
		defer func() {
			// Don't remove for manual inspection
		}()

		err := ExtractQRCodeFromJPEG(testImagePath, tempQRPath)
		if err != nil {
			t.Errorf("ExtractQRCodeFromJPEG failed: %v", err)
			return
		}

		// Read the extracted PNG and analyze its content
		t.Logf("Extracted QR saved to: %s", tempQRPath)
		t.Logf("Check the image manually with: eog %s", tempQRPath)
	})
}

func TestBitstreamPattern(t *testing.T) {
	// Create a test bitstream with known pattern
	size := 128
	bitstream := make([]byte, size*size/8)

	t.Run("CreateTestPattern", func(t *testing.T) {
		// Set some specific patterns
		// QR codes typically have finder patterns in corners

		// Top-left finder pattern simulation (very simplified)
		// Set some bits in the top-left area
		for i := 0; i < 10; i++ {
			bitstream[i] = 0xFF // All black
		}

		// Create test image
		img, err := ConvertBitstreamToQRImage(bitstream, size)
		if err != nil {
			t.Errorf("ConvertBitstreamToQRImage failed: %v", err)
			return
		}

		// Check first few pixels
		bounds := img.Bounds()
		t.Logf("Image bounds: %v", bounds)

		// Check pixel values in top-left corner
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				// Get gray value
				gray := img.(*image.Gray).GrayAt(x, y)
				if y == 0 && x < 5 {
					t.Logf("Pixel (%d,%d): %d", x, y, gray.Y)
				}
			}
		}
	})
}

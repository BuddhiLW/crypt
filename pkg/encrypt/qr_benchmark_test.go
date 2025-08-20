package encrypt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/skip2/go-qrcode"
)

// TestQRCapacityLimits tests the maximum payload for each ECC level
func TestQRCapacityLimits(t *testing.T) {
	levels := []qrcode.RecoveryLevel{qrcode.Highest, qrcode.High, qrcode.Medium, qrcode.Low}
	levelNames := []string{"Highest", "High", "Medium", "Low"}

	t.Log("QR Code Capacity Test Results:")
	t.Log("==============================")

	for i, level := range levels {
		t.Run(levelNames[i], func(t *testing.T) {
			// Binary search for maximum capacity with alphanumeric data
			maxSize := findMaxQRCapacity(t, level, generateAlphanumericData)
			t.Logf("%s ECC - Max alphanumeric bytes: %d", levelNames[i], maxSize)

			// Test with base64-like data (more realistic for encrypted payloads)
			maxBase64Size := findMaxQRCapacity(t, level, generateBase64Data)
			t.Logf("%s ECC - Max base64-like bytes: %d", levelNames[i], maxBase64Size)

			// Test with binary-like data (worst case)
			maxBinarySize := findMaxQRCapacity(t, level, generateBinaryData)
			t.Logf("%s ECC - Max binary-like bytes: %d", levelNames[i], maxBinarySize)

			// Verify the found limits work
			testData := generateBase64Data(maxBase64Size)
			_, err := qrcode.Encode(testData, level, 256)
			if err != nil {
				t.Errorf("Failed to encode at reported max size %d: %v", maxBase64Size, err)
			}

			// Verify one byte over fails
			if maxBase64Size > 0 {
				testData = generateBase64Data(maxBase64Size + 1)
				_, err = qrcode.Encode(testData, level, 256)
				if err == nil {
					t.Errorf("Should have failed to encode at size %d", maxBase64Size+1)
				}
			}
		})
	}
}

// BenchmarkQREncoding benchmarks QR encoding at different sizes and ECC levels
func BenchmarkQREncoding(b *testing.B) {
	levels := []qrcode.RecoveryLevel{qrcode.Highest, qrcode.High, qrcode.Medium, qrcode.Low}
	levelNames := []string{"Highest", "High", "Medium", "Low"}
	sizes := []int{100, 500, 1000, 2000}

	for i, level := range levels {
		for _, size := range sizes {
			b.Run(fmt.Sprintf("%s-%dbytes", levelNames[i], size), func(b *testing.B) {
				data := generateBase64Data(size)
				b.ResetTimer()

				for n := 0; n < b.N; n++ {
					_, err := qrcode.Encode(data, level, 256)
					if err != nil {
						b.Skip("Size too large for this ECC level")
					}
				}
			})
		}
	}
}

// TestQRFallback tests the ECC fallback mechanism
func TestQRFallback(t *testing.T) {
	// Test data that's too large for Highest but fits in Low
	largeData := generateBase64Data(2500) // Should be too big for Highest

	_, level, err := EncodeQRCodeWithFallback(largeData, 256)
	if err != nil {
		t.Fatalf("Fallback encoding failed: %v", err)
	}

	if level == qrcode.Highest {
		t.Error("Expected ECC level to fall back from Highest")
	}

	t.Logf("Large payload fell back to ECC level: %v", level)
}

// TestEmbedCapacity tests the DCT embedding capacity calculation
func TestEmbedCapacity(t *testing.T) {
	// Test with a small payload that should work
	smallData := "Hello, World!"
	encrypted, err := EncryptMessage(smallData, "mysecurepassword123")
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Generate QR and bitstream
	qrBytes, _, err := EncodeQRCodeWithFallback(encrypted, 256)
	if err != nil {
		t.Fatalf("QR encoding failed: %v", err)
	}

	bitstream, err := ExtractBitstreamFromPNG(qrBytes)
	if err != nil {
		t.Fatalf("Bitstream extraction failed: %v", err)
	}

	t.Logf("Small payload QR bitstream: %d bytes (%d bits)", len(bitstream), len(bitstream)*8)

	// This would test actual DCT embedding if we had a test image
	// For now, we'll just verify the bitstream is reasonable
	if len(bitstream) == 0 {
		t.Error("Bitstream should not be empty")
	}

	// QR codes are typically square, so bitstream should be around 256*256/8 bytes
	expectedSize := (256 * 256) / 8 // 8KB for 256x256 QR
	if len(bitstream) > expectedSize*2 {
		t.Errorf("Bitstream seems too large: %d bytes (expected around %d)", len(bitstream), expectedSize)
	}
}

// Helper functions

func findMaxQRCapacity(t *testing.T, level qrcode.RecoveryLevel, dataGenerator func(int) string) int {
	low, high := 1, 10000
	maxSize := 0

	for low <= high {
		mid := (low + high) / 2
		testData := dataGenerator(mid)

		_, err := qrcode.Encode(testData, level, 256)
		if err == nil {
			maxSize = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return maxSize
}

func generateAlphanumericData(size int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}

func generateBase64Data(size int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}

func generateBinaryData(size int) string {
	// Simulate worst-case binary data as string
	return strings.Repeat("\\x00\\xFF", size/8)
}

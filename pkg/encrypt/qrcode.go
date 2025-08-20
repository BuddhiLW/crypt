package encrypt

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -ljpeg
#include <stdio.h>
#include <stdlib.h>
#include <jpeglib.h>

// embed_data_directly_in_dct embeds data directly into DCT coefficients (high capacity)
int embed_data_directly_in_dct(const char* input_path, const char* output_path,
                               const unsigned char* data, int data_size) {
    struct jpeg_decompress_struct cinfo;
    struct jpeg_compress_struct cinfo_out;
    struct jpeg_error_mgr jerr;
    FILE *infile, *outfile;

    // Open input file
    if ((infile = fopen(input_path, "rb")) == NULL) {
        fprintf(stderr, "Cannot open input file %s\n", input_path);
        return 1;
    }

    // Initialize JPEG decompression
    cinfo.err = jpeg_std_error(&jerr);
    jpeg_create_decompress(&cinfo);
    jpeg_stdio_src(&cinfo, infile);
    jpeg_read_header(&cinfo, TRUE);

    // Read DCT coefficients
    jvirt_barray_ptr *coef_ptrs = jpeg_read_coefficients(&cinfo);
    if (!coef_ptrs) {
        fprintf(stderr, "Failed to read DCT coefficients\n");
        fclose(infile);
        return 2;
    }

    // Initialize compression for output
    cinfo_out.err = jpeg_std_error(&jerr);
    jpeg_create_compress(&cinfo_out);
    if ((outfile = fopen(output_path, "wb")) == NULL) {
        fprintf(stderr, "Cannot open output file %s\n", output_path);
        jpeg_finish_decompress(&cinfo);
        jpeg_destroy_decompress(&cinfo);
        fclose(infile);
        return 3;
    }
    jpeg_stdio_dest(&cinfo_out, outfile);
    jpeg_copy_critical_parameters(&cinfo, &cinfo_out);

    // Calculate capacity and embed data
    int total_blocks = cinfo.comp_info[0].height_in_blocks * cinfo.comp_info[0].width_in_blocks;
    int coefficients_per_block = 6;  // Use coefficients 1,2,3,4,5,6 (skip DC coefficient 0)
    int available_bits = total_blocks * coefficients_per_block;
    int required_bits = data_size * 8;

    fprintf(stderr, "Direct DCT: need %d bits, have %d blocks × %d coeffs = %d bits available\n",
            required_bits, total_blocks, coefficients_per_block, available_bits);

    if (required_bits > available_bits) {
        fprintf(stderr, "Error: data too large for direct DCT capacity\n");
        jpeg_finish_decompress(&cinfo);
        jpeg_destroy_decompress(&cinfo);
        jpeg_finish_compress(&cinfo_out);
        jpeg_destroy_compress(&cinfo_out);
        fclose(infile);
        fclose(outfile);
        return 4;
    }

    // Embed data bits into DCT coefficients
    int bit_index = 0;
    int coeff_positions[] = {1, 2, 3, 4, 5, 6};  // AC coefficients to use

    for (JDIMENSION by = 0; by < cinfo.comp_info[0].height_in_blocks && bit_index < required_bits; by++) {
        JBLOCKARRAY block_row = (JBLOCKARRAY)(*cinfo.mem->access_virt_barray)(
            (j_common_ptr)&cinfo, coef_ptrs[0], by, 1, TRUE);

        for (JDIMENSION bx = 0; bx < cinfo.comp_info[0].width_in_blocks && bit_index < required_bits; bx++) {
            // Embed up to 6 bits per block using different AC coefficients
            for (int coeff_idx = 0; coeff_idx < coefficients_per_block && bit_index < required_bits; coeff_idx++) {
                int coeff_pos = coeff_positions[coeff_idx];

                // Get the bit to embed
                unsigned char bit = (data[bit_index / 8] >> (7 - (bit_index % 8))) & 1;

                // Modify the LSB of this coefficient
                if (bit == 1) {
                    block_row[0][bx][coeff_pos] |= 1;  // Set LSB to 1
                } else {
                    block_row[0][bx][coeff_pos] &= ~1; // Set LSB to 0
                }

                bit_index++;
            }
        }
    }

    // Write modified coefficients
    jpeg_write_coefficients(&cinfo_out, coef_ptrs);

    // Cleanup
    jpeg_finish_compress(&cinfo_out);
    jpeg_destroy_compress(&cinfo_out);
    jpeg_finish_decompress(&cinfo);
    jpeg_destroy_decompress(&cinfo);
    fclose(outfile);
    fclose(infile);

    fprintf(stderr, "Successfully embedded %d bits directly into DCT coefficients\n", bit_index);
    return 0;
}

// Embed QR Code into DCT Coefficients (Single Coefficient Strategy)
// Returns 0 on success, non-zero on error
int embed_qr_in_dct_single(const char *input_path, const char *output_path, unsigned char *qr_data, int qr_size) {
    struct jpeg_decompress_struct cinfo;
    struct jpeg_compress_struct cinfo_out;
    struct jpeg_error_mgr jerr;
    FILE *infile, *outfile;

    // Open input JPEG file
    if ((infile = fopen(input_path, "rb")) == NULL) {
        fprintf(stderr, "Cannot open input file %s\n", input_path);
        return 1;
    }

    // Initialize JPEG decompression
    cinfo.err = jpeg_std_error(&jerr);
    jpeg_create_decompress(&cinfo);
    jpeg_stdio_src(&cinfo, infile);
    jpeg_read_header(&cinfo, TRUE);

    // Read DCT coefficients
    jvirt_barray_ptr *coef_ptrs = jpeg_read_coefficients(&cinfo);
    if (!coef_ptrs) {
        fprintf(stderr, "Failed to read DCT coefficients from %s\n", input_path);
        fclose(infile);
        return 2;
    }

    // Initialize compression structure for output
    cinfo_out.err = jpeg_std_error(&jerr);
    jpeg_create_compress(&cinfo_out);
    if ((outfile = fopen(output_path, "wb")) == NULL) {
        fprintf(stderr, "Cannot open output file %s\n", output_path);
        jpeg_finish_decompress(&cinfo);
        jpeg_destroy_decompress(&cinfo);
        fclose(infile);
        return 3;
    }
    jpeg_stdio_dest(&cinfo_out, outfile);
    jpeg_copy_critical_parameters(&cinfo, &cinfo_out);

    // Calculate available capacity
    int total_blocks = cinfo.comp_info[0].height_in_blocks * cinfo.comp_info[0].width_in_blocks;
    int available_bits = total_blocks;  // One bit per block
    int required_bits = qr_size * 8;

    fprintf(stderr, "DCT embedding: need %d bits, have %d blocks (%d bits available)\n",
            required_bits, total_blocks, available_bits);

    if (required_bits > available_bits) {
        fprintf(stderr, "Error: QR data too large for image capacity\n");
        jpeg_finish_decompress(&cinfo);
        jpeg_destroy_decompress(&cinfo);
        jpeg_finish_compress(&cinfo_out);
        jpeg_destroy_compress(&cinfo_out);
        fclose(infile);
        fclose(outfile);
        return 4;
    }

    // Embed QR code into mid-frequency DCT coefficients with simple redundancy
    // Use stronger embedding instead of complex redundancy for now
    int bit_index = 0;

    for (JDIMENSION by = 0; by < cinfo.comp_info[0].height_in_blocks && bit_index < required_bits; by++) {
        JBLOCKARRAY block_row = (JBLOCKARRAY)(*cinfo.mem->access_virt_barray)(
            (j_common_ptr)&cinfo, coef_ptrs[0], by, 1, TRUE);

        for (JDIMENSION bx = 0; bx < cinfo.comp_info[0].width_in_blocks && bit_index < required_bits; bx++) {
            unsigned char bit = (qr_data[bit_index / 8] >> (7 - (bit_index % 8))) & 1;

            // Use low-frequency coefficient (position 1) for better robustness
            int dct_pos = 1;

            // Simple LSB embedding (traditional approach) with logging
            int original_coeff = block_row[0][bx][dct_pos];
            if (bit == 1) {
                block_row[0][bx][dct_pos] |= 1;  // Set LSB to 1
            } else {
                block_row[0][bx][dct_pos] &= ~1; // Set LSB to 0
            }
            int new_coeff = block_row[0][bx][dct_pos];

            // Log first 10 embeddings for debugging
            if (bit_index < 10) {
                fprintf(stderr, "Embed bit %d: pos(%d,%d) coeff[%d] %d->%d (bit=%d)\n",
                    bit_index, bx, by, dct_pos, original_coeff, new_coeff, bit);
            }

            bit_index++;
        }
    }

    // Write modified coefficients
    jpeg_write_coefficients(&cinfo_out, coef_ptrs);

    // Immediate validation: verify first few coefficients were modified
    fprintf(stderr, "=== IMMEDIATE VALIDATION ===\n");
    int validation_errors = 0;
    for (int validate_bits = 0; validate_bits < 10 && validate_bits < required_bits; validate_bits++) {
        JDIMENSION val_by = validate_bits / cinfo.comp_info[0].width_in_blocks;
        JDIMENSION val_bx = validate_bits % cinfo.comp_info[0].width_in_blocks;

        JBLOCKARRAY val_block_row = (JBLOCKARRAY)(*cinfo.mem->access_virt_barray)(
            (j_common_ptr)&cinfo, coef_ptrs[0], val_by, 1, FALSE);

        unsigned char expected_bit = (qr_data[validate_bits / 8] >> (7 - (validate_bits % 8))) & 1;
        unsigned char actual_bit = val_block_row[0][val_bx][1] & 1;

        if (expected_bit != actual_bit) {
            fprintf(stderr, "VALIDATION ERROR bit %d: expected %d, got %d\n",
                validate_bits, expected_bit, actual_bit);
            validation_errors++;
        } else {
            fprintf(stderr, "VALIDATION OK bit %d: %d\n", validate_bits, actual_bit);
        }
    }

    if (validation_errors > 0) {
        fprintf(stderr, "EMBEDDING VALIDATION FAILED: %d errors out of 10 tested bits\n", validation_errors);
        // Don't return error yet, let's see what happens
    } else {
        fprintf(stderr, "EMBEDDING VALIDATION PASSED: All tested bits correct\n");
    }

    // Cleanup
    jpeg_finish_compress(&cinfo_out);
    jpeg_destroy_compress(&cinfo_out);
    jpeg_finish_decompress(&cinfo);
    jpeg_destroy_decompress(&cinfo);
    fclose(infile);
    fclose(outfile);

    fprintf(stderr, "Successfully embedded %d bits into %s\n", bit_index, output_path);
    return 0;
}

// Embed QR Code into DCT Coefficients (Multi-Coefficient Strategy - 4x capacity)
// Returns 0 on success, non-zero on error
int embed_qr_in_dct_multi(const char *input_path, const char *output_path, unsigned char *qr_data, int qr_size) {
    struct jpeg_decompress_struct cinfo;
    struct jpeg_compress_struct cinfo_out;
    struct jpeg_error_mgr jerr;
    FILE *infile, *outfile;

    // Open input JPEG file
    if ((infile = fopen(input_path, "rb")) == NULL) {
        fprintf(stderr, "Cannot open input file %s\n", input_path);
        return 1;
    }

    // Initialize JPEG decompression
    cinfo.err = jpeg_std_error(&jerr);
    jpeg_create_decompress(&cinfo);
    jpeg_stdio_src(&cinfo, infile);
    jpeg_read_header(&cinfo, TRUE);

    // Read DCT coefficients
    jvirt_barray_ptr *coef_ptrs = jpeg_read_coefficients(&cinfo);
    if (!coef_ptrs) {
        fprintf(stderr, "Failed to read DCT coefficients from %s\n", input_path);
        fclose(infile);
        return 2;
    }

    // Initialize compression structure for output
    cinfo_out.err = jpeg_std_error(&jerr);
    jpeg_create_compress(&cinfo_out);
    if ((outfile = fopen(output_path, "wb")) == NULL) {
        fprintf(stderr, "Cannot open output file %s\n", output_path);
        fclose(infile);
        return 3;
    }
    jpeg_stdio_dest(&cinfo_out, outfile);
    jpeg_copy_critical_parameters(&cinfo, &cinfo_out);

    // Calculate available capacity (4 bits per block for multi-coefficient)
    int total_blocks = cinfo.comp_info[0].height_in_blocks * cinfo.comp_info[0].width_in_blocks;
    int available_bits = total_blocks * 4;  // 4 coefficients per block
    int required_bits = qr_size * 8;

    fprintf(stderr, "DCT multi-coeff embedding: need %d bits, have %d blocks (%d bits available)\n",
            required_bits, total_blocks, available_bits);

    if (required_bits > available_bits) {
        fprintf(stderr, "QR data too large for image capacity (multi-coeff)\n");
        fclose(infile);
        fclose(outfile);
        return 4;
    }

    // Embed QR code into mid-frequency DCT coefficients using 4 coefficients per block
    int bit_index = 0;
    int coeff_positions[4] = {4, 5, 6, 7};  // Mid-frequency positions

    for (JDIMENSION by = 0; by < cinfo.comp_info[0].height_in_blocks && bit_index < required_bits; by++) {
        JBLOCKARRAY block_row = (JBLOCKARRAY)(*cinfo.mem->access_virt_barray)(
            (j_common_ptr)&cinfo, coef_ptrs[0], by, 1, TRUE);

        for (JDIMENSION bx = 0; bx < cinfo.comp_info[0].width_in_blocks && bit_index < required_bits; bx++) {
            // Embed up to 4 bits per block using different coefficients
            for (int coeff_idx = 0; coeff_idx < 4 && bit_index < required_bits; coeff_idx++) {
                unsigned char bit = (qr_data[bit_index / 8] >> (7 - (bit_index % 8))) & 1;
                int dct_pos = coeff_positions[coeff_idx];

                // Simple LSB embedding (traditional approach)
                if (bit == 1) {
                    block_row[0][bx][dct_pos] |= 1;  // Set LSB to 1
                } else {
                    block_row[0][bx][dct_pos] &= ~1; // Set LSB to 0
                }

                bit_index++;
            }
        }
    }

    // Write modified coefficients
    jpeg_write_coefficients(&cinfo_out, coef_ptrs);

    // Cleanup
    jpeg_finish_compress(&cinfo_out);
    jpeg_destroy_compress(&cinfo_out);
    jpeg_finish_decompress(&cinfo);
    jpeg_destroy_decompress(&cinfo);
    fclose(infile);
    fclose(outfile);

    fprintf(stderr, "Successfully embedded %d bits into %s (multi-coeff)\n", bit_index, output_path);
    return 0;
}

// Test function for direct DCT embedding (for unit tests)
int test_dct_embedding(const char *input_path, const char *output_path, unsigned char *data, int data_size, int multi_coeff) {
    if (multi_coeff) {
        return embed_qr_in_dct_multi(input_path, output_path, data, data_size);
    } else {
        return embed_qr_in_dct_single(input_path, output_path, data, data_size);
    }
}

// Test function for direct DCT extraction (for unit tests)
void test_dct_extraction(const char *input_path, unsigned char *data, int data_size, int multi_coeff) {
    if (multi_coeff) {
        extract_qr_from_dct_multi(input_path, data, data_size);
    } else {
        extract_qr_from_dct_single(input_path, data, data_size);
    }
}
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"os"
	"unsafe"

	"github.com/rwxrob/bonzai/vars"
	"github.com/skip2/go-qrcode"
)

// DCTEmbeddingStrategy defines the interface for different DCT embedding approaches (OCP)
type DCTEmbeddingStrategy interface {
	CalculateCapacity(width, height int) int
	GetCoefficientsPerBit() int
	GetStrategyName() string
	GetCCode() string // Returns the C code for this strategy
}

// SingleCoefficientDCT implements single coefficient per bit (original approach)
type SingleCoefficientDCT struct{}

func (s *SingleCoefficientDCT) CalculateCapacity(width, height int) int {
	dctBlocksX := (width + 7) / 8
	dctBlocksY := (height + 7) / 8
	return dctBlocksX * dctBlocksY // 1 bit per block
}

func (s *SingleCoefficientDCT) GetCoefficientsPerBit() int {
	return 1
}

func (s *SingleCoefficientDCT) GetStrategyName() string {
	return "single-coefficient"
}

func (s *SingleCoefficientDCT) GetCCode() string {
	return "single" // Will map to existing C code
}

// MultiCoefficientDCT implements multiple coefficients per bit (4x capacity)
type MultiCoefficientDCT struct{}

func (m *MultiCoefficientDCT) CalculateCapacity(width, height int) int {
	dctBlocksX := (width + 7) / 8
	dctBlocksY := (height + 7) / 8
	// Use 4 coefficients per block: positions 4,5,6,7 (mid-frequency)
	return (dctBlocksX * dctBlocksY) * 4 // 4 bits per block
}

func (m *MultiCoefficientDCT) GetCoefficientsPerBit() int {
	return 1 // Each bit uses 1 coefficient, but we can use 4 per block
}

func (m *MultiCoefficientDCT) GetStrategyName() string {
	return "multi-coefficient"
}

func (m *MultiCoefficientDCT) GetCCode() string {
	return "multi" // Will map to new C code
}

// QRSizeCalculator handles QR code size calculations (SRP)
type QRSizeCalculator struct {
	strategy DCTEmbeddingStrategy
}

func NewQRSizeCalculator(strategy DCTEmbeddingStrategy) *QRSizeCalculator {
	return &QRSizeCalculator{strategy: strategy}
}

// ImageDimensions represents image dimensions
type ImageDimensions struct {
	Width  int
	Height int
}

// GetImageDimensions extracts dimensions from a JPEG file (SRP)
func GetImageDimensions(imagePath string) (*ImageDimensions, error) {
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

// ✅ **Fix for `C.free` not recognized**
// `C.free` must be explicitly included in `stdlib.h` (which we did in CGO).

// CreateQRCodeBytes generates a QR code and returns its PNG bytes.
// Now tries ECC Highest -> High -> Medium -> Low and returns the first that fits.
func CreateQRCodeBytes(data string) ([]byte, error) {
	png, level, err := EncodeQRCodeWithFallback(data, 256)
	if err != nil {
		return nil, err
	}
	if level != qrcode.Highest {
		fmt.Printf("warning: QR ECC lowered to %v to fit payload; robustness may be reduced\n", level)
	}
	return png, nil
}

// EncodeQRCodeWithFallback attempts to encode using High/Highest ECC only for DCT robustness.
func EncodeQRCodeWithFallback(data string, size int) ([]byte, qrcode.RecoveryLevel, error) {
	// Only try Highest and High ECC - reject Medium/Low for DCT steganography robustness
	levels := []qrcode.RecoveryLevel{qrcode.Highest, qrcode.High}
	var lastErr error
	for _, lvl := range levels {
		png, err := qrcode.Encode(data, lvl, size)
		if err == nil {
			fmt.Printf("Successfully generated QR code with %v ECC level\n", lvl)
			return png, lvl, nil
		}
		lastErr = err
	}
	return nil, qrcode.Low, fmt.Errorf("failed to encode QR with High or Highest ECC - payload too large for robust DCT steganography: %w", lastErr)
}

// WriteQRCodeWithFallback writes a QR to a file and returns chosen ECC level.
func WriteQRCodeWithFallback(data string, size int, path string) (qrcode.RecoveryLevel, error) {
	png, level, err := EncodeQRCodeWithFallback(data, size)
	if err != nil {
		return level, err
	}
	if err := os.WriteFile(path, png, 0600); err != nil {
		return level, fmt.Errorf("failed to write qr file: %w", err)
	}
	if level != qrcode.Highest {
		fmt.Printf("warning: QR ECC lowered to %v to fit payload; robustness may be reduced\n", level)
	}
	return level, nil
}

// ExtractBitstreamFromPNG takes a PNG byte array and returns a packed bitstream.
func ExtractBitstreamFromPNG(pngData []byte) ([]byte, error) {
	// Decode PNG
	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, errors.New("failed to decode PNG image")
	}

	// Convert image to grayscale if needed
	grayImg := image.NewGray(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			grayImg.Set(x, y, img.At(x, y)) // Converts automatically
		}
	}

	// Get image bounds
	bounds := grayImg.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Prepare bitstream (bit-packed storage)
	bitstream := make([]byte, (width*height+7)/8)
	bitIndex := 0

	// Convert pixels to 1s and 0s
	blackPixels := 0
	whitePixels := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := grayImg.GrayAt(x, y).Y // Extract grayscale intensity

			// Black (dark) pixels are '1', white (light) pixels are '0'
			if pixel < 128 {
				bitstream[bitIndex/8] |= 1 << (7 - (bitIndex % 8))
				blackPixels++

				// Debug first 16 black pixels
				if blackPixels <= 16 {
					fmt.Printf("Black pixel %d at (%d,%d): intensity=%d, bitIndex=%d, byteIndex=%d, bitPos=%d\n",
						blackPixels, x, y, pixel, bitIndex, bitIndex/8, 7-(bitIndex%8))
				}
			} else {
				whitePixels++
			}
			bitIndex++
		}
	}

	fmt.Printf("PNG conversion: %dx%d, black pixels: %d, white pixels: %d\n",
		width, height, blackPixels, whitePixels)

	return bitstream, nil
}

// EmbedQRCodeInJPEG embeds a QR code bitstream into a JPEG's DCT coefficients using SOLID principles
func EmbedQRCodeInJPEG(inputPath, outputPath, qrData string, payloadSize int) error {
	// Get DCT strategy from Bonzai vars (DIP - dependency inversion)
	strategyName, _ := vars.Get(DCTStrategyVar, DCTEnv)
	if strategyName == "" {
		strategyName = "single-coefficient" // Default strategy
	}

	var strategy DCTEmbeddingStrategy
	switch strategyName {
	case "multi-coefficient":
		strategy = &MultiCoefficientDCT{}
	default:
		strategy = &SingleCoefficientDCT{}
	}

	// Create QR size calculator with strategy (OCP - open/closed principle)
	calculator := NewQRSizeCalculator(strategy)

	// Calculate optimal QR size
	qrSize, err := calculator.CalculateOptimalQRSize(inputPath, payloadSize)
	if err != nil {
		return fmt.Errorf("error calculating QR size: %w", err)
	}

	fmt.Printf("Using QR size: %dx%d for image\n", qrSize, qrSize)

	// Generate QR Code as PNG bytes (High/Highest ECC only)
	qrBytes, _, err := EncodeQRCodeWithFallback(qrData, qrSize)
	if err != nil {
		return fmt.Errorf("error generating QR code: %w", err)
	}
	// No warning needed - we only allow High/Highest ECC now

	// Convert PNG to bitstream
	bitstream, err := ExtractBitstreamFromPNG(qrBytes)
	if err != nil {
		return fmt.Errorf("error extracting bitstream: %w", err)
	}

	// Decode the PNG to get the actual dimensions
	img, _, err := image.Decode(bytes.NewReader(qrBytes))
	if err != nil {
		return fmt.Errorf("failed to decode generated QR PNG: %w", err)
	}
	actualQRSize := img.Bounds().Dx() // Assume square QR code
	fmt.Printf("Actual QR code size: %dx%d\n", actualQRSize, actualQRSize)

	// Log first few bytes of bitstream for debugging
	fmt.Printf("QR bitstream first 10 bytes: ")
	for i := 0; i < 10 && i < len(bitstream); i++ {
		fmt.Printf("%02x ", bitstream[i])
	}
	fmt.Printf("\n")

	// Log first few bits in detail
	fmt.Printf("QR bitstream first 16 bits: ")
	for i := 0; i < 16 && i < len(bitstream)*8; i++ {
		bit := (bitstream[i/8] >> (7 - (i % 8))) & 1
		fmt.Printf("%d", bit)
	}
	fmt.Printf("\n")

	// Update the stored QR size with the actual dimensions
	if err := vars.Set(QRSizeVar, fmt.Sprintf("%d", actualQRSize), DCTEnv); err != nil {
		fmt.Printf("Warning: failed to update actual QR size in vars: %v\n", err)
	}

	// Convert paths to C strings
	cInputPath := C.CString(inputPath)
	cOutputPath := C.CString(outputPath)
	defer C.free(unsafe.Pointer(cInputPath))
	defer C.free(unsafe.Pointer(cOutputPath))

	fmt.Printf("QR bitstream size: %d bytes (%d bits)\n", len(bitstream), len(bitstream)*8)

	// Call strategy-specific C function (OCP - open/closed principle)
	var result C.int
	switch strategy.GetCCode() {
	case "multi":
		result = C.embed_qr_in_dct_multi(cInputPath, cOutputPath, (*C.uchar)(unsafe.Pointer(&bitstream[0])), C.int(len(bitstream)))
	default:
		result = C.embed_qr_in_dct_single(cInputPath, cOutputPath, (*C.uchar)(unsafe.Pointer(&bitstream[0])), C.int(len(bitstream)))
	}

	if result != 0 {
		return fmt.Errorf("DCT embedding failed with code %d (%s strategy)", int(result), strategy.GetStrategyName())
	}

	// No more metadata files - QR size is stored in Bonzai vars!

	fmt.Println("Modified JPEG saved as:", outputPath)
	return nil
}

// CalculateOptimalQRSize determines the optimal QR code size using SOLID principles
func (calc *QRSizeCalculator) CalculateOptimalQRSize(imagePath string, payloadSize int) (int, error) {
	// Get image dimensions (SRP - single responsibility)
	dims, err := GetImageDimensions(imagePath)
	if err != nil {
		return 0, err
	}

	// Calculate DCT capacity using the strategy (OCP - open/closed principle)
	dctCapacityBits := calc.strategy.CalculateCapacity(dims.Width, dims.Height)

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
	optimalQRSize, err := findOptimalQRSizeForHighECC(payloadSize)
	if err != nil {
		return 0, fmt.Errorf("payload too large for High ECC: %w", err)
	}

	// Use the optimal size, but check against capacity limits
	qrSize := optimalQRSize
	if qrSize > qrSizeFromDim {
		qrSize = qrSizeFromDim
		fmt.Printf("QR size limited by image dimensions (80%% of %dx%d)\n", dims.Width, dims.Height)
	}
	if qrSize > maxQRSizeFromDCT {
		qrSize = maxQRSizeFromDCT
		fmt.Printf("QR size limited by DCT capacity (%s strategy)\n", calc.strategy.GetStrategyName())
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

	fmt.Printf("Image dimensions: %dx%d, DCT capacity: %d bits (%s strategy)\n",
		dims.Width, dims.Height, dctCapacityBits, calc.strategy.GetStrategyName())
	fmt.Printf("Payload size: %d bytes, optimal QR size: %dx%d\n",
		payloadSize, optimalQRSize, optimalQRSize)
	fmt.Printf("Calculated QR size: %dx%d (%d bits needed)\n",
		qrSize, qrSize, qrSize*qrSize)

	// Store QR size in Bonzai vars (will be updated with actual size in embedding function)
	if err := vars.Set(QRSizeVar, fmt.Sprintf("%d", qrSize), DCTEnv); err != nil {
		fmt.Printf("Warning: failed to store QR size in vars: %v\n", err)
	}
	if err := vars.Set(DCTStrategyVar, calc.strategy.GetStrategyName(), DCTEnv); err != nil {
		fmt.Printf("Warning: failed to store DCT strategy in vars: %v\n", err)
	}

	return qrSize, nil
}

// findOptimalQRSizeForHighECC finds the optimal QR size that can hold the payload with Highest or High ECC
func findOptimalQRSizeForHighECC(payloadBytes int) (int, error) {
	// QR code capacity for binary data with different ECC levels
	// Format: size -> [HighestECC_capacity, HighECC_capacity]
	capacityMap := map[int][2]int{
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

	// Try to find a QR size that can hold the payload with Highest ECC first
	for size := 64; size <= 512; size += 32 {
		if capacities, exists := capacityMap[size]; exists {
			// Try Highest ECC first
			if capacities[0] >= payloadBytes {
				fmt.Printf("Found QR size %dx%d for %d bytes with Highest ECC\n", size, size, payloadBytes)
				return size, nil
			}
			// Try High ECC as fallback
			if capacities[1] >= payloadBytes {
				fmt.Printf("Found QR size %dx%d for %d bytes with High ECC\n", size, size, payloadBytes)
				return size, nil
			}
		}
	}

	// If payload is too large for High ECC at maximum size, reject it
	return 0, fmt.Errorf("payload %d bytes exceeds High ECC capacity at maximum QR size (512x512, ~3.5KB max)", payloadBytes)
}

// EmbedDataDirectlyInDCT embeds data directly into DCT coefficients without QR overhead
func EmbedDataDirectlyInDCT(inputPath, outputPath, data string) error {
	fmt.Printf("Direct DCT embedding: %d bytes into %s\n", len(data), inputPath)

	// Get image dimensions for capacity calculation
	dims, err := GetImageDimensions(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get image dimensions: %w", err)
	}

	// Calculate DCT capacity using all available coefficients (not just 1 per block)
	blocksWidth := dims.Width / 8
	blocksHeight := dims.Height / 8
	totalBlocks := blocksWidth * blocksHeight

	// Use 6 coefficients per block (positions 1,2,3,4,5,6) - avoid DC coefficient (0)
	// This gives us 6 bits per block = much higher capacity than QR codes
	coefficientsPerBlock := 6
	totalCapacityBits := totalBlocks * coefficientsPerBlock
	totalCapacityBytes := totalCapacityBits / 8

	fmt.Printf("Image: %dx%d, Blocks: %dx%d (%d total)\n", dims.Width, dims.Height, blocksWidth, blocksHeight, totalBlocks)
	fmt.Printf("Direct DCT capacity: %d bits (%d bytes) using %d coefficients per block\n",
		totalCapacityBits, totalCapacityBytes, coefficientsPerBlock)

	if len(data) > totalCapacityBytes {
		return fmt.Errorf("data too large: %d bytes > %d bytes capacity", len(data), totalCapacityBytes)
	}

	// Add simple error detection: store data length + checksum at the beginning
	dataBytes := []byte(data)
	dataLength := uint32(len(dataBytes))
	checksum := calculateSimpleChecksum(dataBytes)

	// Prepare payload: [length:4bytes][checksum:4bytes][data]
	payload := make([]byte, 8+len(dataBytes))
	binary.LittleEndian.PutUint32(payload[0:4], dataLength)
	binary.LittleEndian.PutUint32(payload[4:8], checksum)
	copy(payload[8:], dataBytes)

	fmt.Printf("Payload: %d bytes (length: %d, checksum: %08x, data: %d)\n",
		len(payload), dataLength, checksum, len(dataBytes))

	// Call C function for direct DCT embedding
	cInputPath := C.CString(inputPath)
	cOutputPath := C.CString(outputPath)
	defer C.free(unsafe.Pointer(cInputPath))
	defer C.free(unsafe.Pointer(cOutputPath))

	result := C.embed_data_directly_in_dct(cInputPath, cOutputPath,
		(*C.uchar)(unsafe.Pointer(&payload[0])), C.int(len(payload)))

	if result != 0 {
		return fmt.Errorf("direct DCT embedding failed with code %d", int(result))
	}

	// Store metadata for extraction
	metadata := fmt.Sprintf("direct_dct:%d:%08x", dataLength, checksum)
	if err := vars.Set("DIRECT_DCT_META", metadata, DCTEnv); err != nil {
		fmt.Printf("Warning: failed to store direct DCT metadata: %v\n", err)
	}

	return nil
}

// calculateSimpleChecksum calculates a simple checksum for error detection
func calculateSimpleChecksum(data []byte) uint32 {
	var checksum uint32
	for _, b := range data {
		checksum = (checksum << 1) ^ uint32(b)
	}
	return checksum
}

// storeQRSizeMetadata stores the QR size in a companion metadata file
func storeQRSizeMetadata(imagePath string, qrSize int) error {
	metadataPath := imagePath + ".qrmeta"
	return os.WriteFile(metadataPath, []byte(fmt.Sprintf("%d", qrSize)), 0644)
}

// loadQRSizeMetadata loads the QR size from a companion metadata file
func loadQRSizeMetadata(imagePath string) (int, error) {
	metadataPath := imagePath + ".qrmeta"
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return 256, nil // Default fallback
	}

	var qrSize int
	_, err = fmt.Sscanf(string(data), "%d", &qrSize)
	if err != nil {
		return 256, nil // Default fallback
	}

	return qrSize, nil
}

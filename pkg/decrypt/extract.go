package decrypt

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -ljpeg
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <jpeglib.h>

// Extract QR Code from DCT Coefficients (Single Coefficient Strategy)
void extract_qr_from_dct_single(const char *input_path, unsigned char *qr_data, int qr_size) {
    struct jpeg_decompress_struct cinfo;
    struct jpeg_error_mgr jerr;
    FILE *infile;

    // Open input JPEG file
    if ((infile = fopen(input_path, "rb")) == NULL) {
        fprintf(stderr, "Cannot open file %s\n", input_path);
        return;
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
        return;
    }

    // Clear output buffer first (critical fix!)
    memset(qr_data, 0, qr_size);

    // Extract QR code bits from mid-frequency DCT coefficients with stronger detection
    int bit_index = 0;
    for (JDIMENSION by = 0; by < cinfo.comp_info[0].height_in_blocks; by++) {
        JBLOCKARRAY block_row = (JBLOCKARRAY)(*cinfo.mem->access_virt_barray)(
            (j_common_ptr)&cinfo, coef_ptrs[0], by, 1, FALSE);

        for (JDIMENSION bx = 0; bx < cinfo.comp_info[0].width_in_blocks; bx++) {
            if (bit_index >= qr_size * 8) break;  // Stop when enough bits are extracted

            			// Extract LSB from low-frequency coefficient for better robustness
			int dct_pos = 1; // Low-frequency coefficient
            int coeff_value = block_row[0][bx][dct_pos];
            unsigned char bit = coeff_value & 1; // Extract LSB

            // Store bit into output buffer
            if (bit == 1)
                qr_data[bit_index / 8] |= 1 << (7 - (bit_index % 8)); // Set bit
            else
                qr_data[bit_index / 8] &= ~(1 << (7 - (bit_index % 8))); // Clear bit

            // Log first 10 extractions for debugging
            if (bit_index < 10) {
                fprintf(stderr, "Extract bit %d: pos(%d,%d) coeff[%d]=%d LSB=%d\n",
                    bit_index, bx, by, dct_pos, coeff_value, bit);
            }

            bit_index++;
        }
    }

    // Cleanup
    jpeg_finish_decompress(&cinfo);
    jpeg_destroy_decompress(&cinfo);
    fclose(infile);
}

// Extract QR Code from DCT Coefficients (Multi-Coefficient Strategy)
void extract_qr_from_dct_multi(const char *input_path, unsigned char *qr_data, int qr_size) {
    struct jpeg_decompress_struct cinfo;
    struct jpeg_error_mgr jerr;
    FILE *infile;

    // Open input JPEG file
    if ((infile = fopen(input_path, "rb")) == NULL) {
        fprintf(stderr, "Cannot open file %s\n", input_path);
        return;
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
        return;
    }

    // Clear output buffer first (critical fix!)
    memset(qr_data, 0, qr_size);

    // Extract QR code bits from mid-frequency DCT coefficients (multi-coefficient)
    int bit_index = 0;
    int required_bits = qr_size * 8;
    int coeff_positions[4] = {4, 5, 6, 7};  // Same positions used during embedding

    for (JDIMENSION by = 0; by < cinfo.comp_info[0].height_in_blocks; by++) {
        JBLOCKARRAY block_row = (JBLOCKARRAY)(*cinfo.mem->access_virt_barray)(
            (j_common_ptr)&cinfo, coef_ptrs[0], by, 1, FALSE);

        for (JDIMENSION bx = 0; bx < cinfo.comp_info[0].width_in_blocks; bx++) {
            if (bit_index >= required_bits) break;  // Stop when enough bits are extracted

            // Extract up to 4 bits per block from different coefficients
            for (int coeff_idx = 0; coeff_idx < 4 && bit_index < required_bits; coeff_idx++) {
                int dct_pos = coeff_positions[coeff_idx];
                unsigned char bit = block_row[0][bx][dct_pos] & 1; // Extract LSB

                // Store bit into output buffer
                if (bit == 1)
                    qr_data[bit_index / 8] |= 1 << (7 - (bit_index % 8)); // Set bit
                else
                    qr_data[bit_index / 8] &= ~(1 << (7 - (bit_index % 8))); // Clear bit

                bit_index++;
            }
        }
    }

    // Cleanup
    jpeg_finish_decompress(&cinfo);
    jpeg_destroy_decompress(&cinfo);
    fclose(infile);
}
*/
import "C"
import (
	// "bytes"
	// "github.com/skip2/go-qrcode"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"unsafe"

	"github.com/rwxrob/bonzai/vars"
)

const (
	QRSizeVar = `qr-size`

	DCTEnv         = `DCT_ENV`
	DCTStrategyVar = `dct-strategy`
)

// ExtractQRCodeFromJPEG extracts QR code bits from a JPEG's DCT coefficients using SOLID principles
func ExtractQRCodeFromJPEG(inputPath string, outputQRPath string) error {
	// Get QR size from Bonzai vars (DIP - dependency inversion)
	qrSizeStr, _ := vars.Get(QRSizeVar, DCTEnv)
	if qrSizeStr == "" {
		qrSizeStr = "256" // Default size
	}
	qrPixelSize, err := strconv.Atoi(qrSizeStr)
	if err != nil {
		fmt.Printf("Warning: invalid QR size in vars, using default 256: %v\n", err)
		qrPixelSize = 256
	}

	// Get DCT strategy from Bonzai vars
	strategyName, _ := vars.Get(DCTStrategyVar, DCTEnv)
	if strategyName == "" {
		strategyName = "single-coefficient" // Default strategy
	}

	fmt.Printf("Extracting QR code with size: %dx%d (strategy: %s)\n", qrPixelSize, qrPixelSize, strategyName)

	// Get the actual data area from Bonzai vars (this is the critical fix!)
	qrDataAreaStr, _ := vars.Get("QR_DATA_AREA", DCTEnv)
	var qrDataArea int
	if qrDataAreaStr != "" {
		qrDataArea, err = strconv.Atoi(qrDataAreaStr)
		if err != nil {
			fmt.Printf("Warning: invalid QR data area in vars, falling back to data size: %v\n", err)
			qrDataArea = qrPixelSize
		}
	} else {
		// Fallback to pixel size if data area not stored
		qrDataArea = qrPixelSize
		fmt.Printf("Warning: QR data area not found in vars, using pixel size: %dx%d\n", qrDataArea, qrDataArea)
	}

	// Calculate the actual data size based on the data area (not full pixels)
	qrDataSize := qrDataArea * qrDataArea / 8
	fmt.Printf("Using QR data area: %dx%d, calculated data size: %d bytes\n", qrDataArea, qrDataArea, qrDataSize)

	// Prepare buffer to receive QR bitstream - we need full pixel size for C extraction
	// but we'll only use the data area portion for reconstruction
	fullPixelSize := qrPixelSize * qrPixelSize / 8
	qrBitstream := make([]byte, fullPixelSize)

	// Convert input path to C string
	cInputPath := C.CString(inputPath)
	defer C.free(unsafe.Pointer(cInputPath))

	// Extract QR bitstream using strategy-specific CGO function (OCP - open/closed principle)
	fmt.Println("Extracting QR Code from JPEG DCT coefficients...")

	// The C function needs to know how many bits to extract, but we need to ensure
	// we don't overflow our buffer. We'll extract the full pixel area but only
	// use the data area portion.
	fmt.Printf("Extracting %d bits (full pixel area) into %d-byte buffer (data area)\n",
		fullPixelSize*8, qrDataSize)

	switch strategyName {
	case "multi-coefficient":
		C.extract_qr_from_dct_multi(cInputPath, (*C.uchar)(unsafe.Pointer(&qrBitstream[0])), C.int(fullPixelSize))
	default:
		C.extract_qr_from_dct_single(cInputPath, (*C.uchar)(unsafe.Pointer(&qrBitstream[0])), C.int(fullPixelSize))
	}

	// Log extracted bitstream for comparison
	fmt.Printf("Extracted bitstream first 10 bytes: ")
	for i := 0; i < 10 && i < len(qrBitstream); i++ {
		fmt.Printf("%02x ", qrBitstream[i])
	}
	fmt.Printf("\n")

	// Log bytes around where we expect QR data (around byte 109 based on embedding output)
	fmt.Printf("Extracted bitstream bytes 109-119: ")
	for i := 109; i < 120 && i < len(qrBitstream); i++ {
		fmt.Printf("%02x ", qrBitstream[i])
	}
	fmt.Printf("\n")

	// Log first few bits in detail
	fmt.Printf("Extracted bitstream first 16 bits: ")
	for i := 0; i < 16 && i < len(qrBitstream)*8; i++ {
		bit := (qrBitstream[i/8] >> (7 - (i % 8))) & 1
		fmt.Printf("%d", bit)
	}
	fmt.Printf("\n")

	// Convert bitstream to QR image using the full pixel size (not just data area)
	// The full bitstream contains all the QR code data including quiet zones
	img, err := ConvertBitstreamToQRImage(qrBitstream, qrPixelSize)
	if err != nil {
		return fmt.Errorf("failed to reconstruct QR image: %w", err)
	}

	// Save QR code as PNG
	file, err := os.Create(outputQRPath)
	if err != nil {
		return fmt.Errorf("failed to create QR output file: %w", err)
	}
	defer file.Close()
	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode QR PNG: %w", err)
	}

	fmt.Println("Extracted QR Code saved to:", outputQRPath)
	return nil
}

// ConvertBitstreamToQRImage reconstructs a QR code image from bitstream
func ConvertBitstreamToQRImage(bitstream []byte, size int) (image.Image, error) {
	img := image.NewGray(image.Rect(0, 0, size, size))
	bitIndex := 0

	for y := range size {
		for x := range size {
			bit := (bitstream[bitIndex/8] >> (7 - (bitIndex % 8))) & 1
			if bit == 1 {
				img.SetGray(x, y, color.Gray{Y: 0}) // Black
			} else {
				img.SetGray(x, y, color.Gray{Y: 255}) // White
			}
			bitIndex++
		}
	}

	return img, nil
}

// No more metadata files! QR size is now stored in Bonzai vars (SOLID principles)

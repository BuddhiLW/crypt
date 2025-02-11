package decrypt

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -ljpeg
#include <stdio.h>
#include <stdlib.h>
#include <jpeglib.h>

// Extract QR Code from DCT Coefficients
void extract_qr_from_dct(const char *input_path, unsigned char *qr_data, int qr_size) {
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

    // Extract QR code bits from mid-frequency DCT coefficients
    int bit_index = 0;
    for (JDIMENSION by = 0; by < cinfo.comp_info[0].height_in_blocks; by++) {
        JBLOCKARRAY block_row = (JBLOCKARRAY)(*cinfo.mem->access_virt_barray)(
            (j_common_ptr)&cinfo, coef_ptrs[0], by, 1, FALSE);

        for (JDIMENSION bx = 0; bx < cinfo.comp_info[0].width_in_blocks; bx++) {
            if (bit_index >= qr_size * 8) break;  // Stop when enough bits are extracted

            // Extract LSB from mid-frequency coefficient
            int dct_x = 4; // Mid-frequency coefficient
            unsigned char bit = block_row[0][bx][dct_x] & 1; // Extract LSB

            // Store bit into output buffer
            if (bit == 1)
                qr_data[bit_index / 8] |= 1 << (7 - (bit_index % 8)); // Set bit
            else
                qr_data[bit_index / 8] &= ~(1 << (7 - (bit_index % 8))); // Clear bit

            bit_index++;
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
	"unsafe"
)

// ExtractQRCodeFromJPEG extracts QR code bits from a JPEG's DCT coefficients
func ExtractQRCodeFromJPEG(inputPath string, outputQRPath string) error {
	// Prepare buffer to receive QR bitstream
	qrSize := 256 * 256 / 8 // Same as the embedded QR size
	qrBitstream := make([]byte, qrSize)

	// Convert input path to C string
	cInputPath := C.CString(inputPath)
	defer C.free(unsafe.Pointer(cInputPath))

	// Extract QR bitstream using CGO
	fmt.Println("Extracting QR Code from JPEG DCT coefficients...")
	C.extract_qr_from_dct(cInputPath, (*C.uchar)(unsafe.Pointer(&qrBitstream[0])), C.int(qrSize))

	// Convert bitstream to QR image
	img, err := ConvertBitstreamToQRImage(qrBitstream, 256)
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

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
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

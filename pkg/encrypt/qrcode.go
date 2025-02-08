package encrypt

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -ljpeg
#include <stdio.h>
#include <stdlib.h>
#include <jpeglib.h>

// Embed QR Code into DCT Coefficients
void embed_qr_in_dct(const char *input_path, const char *output_path, unsigned char *qr_data, int qr_size) {
    struct jpeg_decompress_struct cinfo;
    struct jpeg_compress_struct cinfo_out;
    struct jpeg_error_mgr jerr;
    FILE *infile, *outfile;

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

    // Initialize compression structure for output
    cinfo_out.err = jpeg_std_error(&jerr);
    jpeg_create_compress(&cinfo_out);
    if ((outfile = fopen(output_path, "wb")) == NULL) {
        fprintf(stderr, "Cannot open file %s\n", output_path);
        fclose(infile);
        return;
    }
    jpeg_stdio_dest(&cinfo_out, outfile);
    jpeg_copy_critical_parameters(&cinfo, &cinfo_out);

    // Embed QR code into mid-frequency DCT coefficients
    int bit_index = 0;
    for (JDIMENSION by = 0; by < cinfo.comp_info[0].height_in_blocks; by++) {
        JBLOCKARRAY block_row = (JBLOCKARRAY)(*cinfo.mem->access_virt_barray)(
            (j_common_ptr)&cinfo, coef_ptrs[0], by, 1, TRUE);

        for (JDIMENSION bx = 0; bx < cinfo.comp_info[0].width_in_blocks; bx++) {
            if (bit_index >= qr_size * 8) break;  // Stop if QR fully embedded

            // Select mid-frequency coefficient
            int dct_x = 4;  // Mid-frequency
            unsigned char bit = (qr_data[bit_index / 8] >> (7 - (bit_index % 8))) & 1;

            // Set LSB to QR bit
            if (bit == 1)
                block_row[0][bx][dct_x] |= 1;  // Set LSB to 1
            else
                block_row[0][bx][dct_x] &= ~1; // Set LSB to 0

            bit_index++;
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
}
*/
import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"unsafe"

	"github.com/skip2/go-qrcode"
)

// ✅ **Fix for `C.free` not recognized**
// `C.free` must be explicitly included in `stdlib.h` (which we did in CGO).

// CreateQRCodeBytes generates a QR code and returns its PNG bytes.
func CreateQRCodeBytes(data string) ([]byte, error) {
	return qrcode.Encode(data, qrcode.Highest, 256)
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
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := grayImg.GrayAt(x, y).Y // Extract grayscale intensity

			// Black (dark) pixels are '1', white (light) pixels are '0'
			if pixel < 128 {
				bitstream[bitIndex/8] |= 1 << (7 - (bitIndex % 8))
			}
			bitIndex++
		}
	}

	return bitstream, nil
}

// EmbedQRCodeInJPEG embeds a QR code bitstream into a JPEG's DCT coefficients
func EmbedQRCodeInJPEG(inputPath, outputPath, qrData string) error {
	// Generate QR Code as PNG bytes
	qrBytes, err := qrcode.Encode(qrData, qrcode.Highest, 256)
	if err != nil {
		return fmt.Errorf("error generating QR code: %w", err)
	}

	// Convert PNG to bitstream
	bitstream, err := ExtractBitstreamFromPNG(qrBytes)
	if err != nil {
		return fmt.Errorf("error extracting bitstream: %w", err)
	}

	// Convert paths to C strings
	cInputPath := C.CString(inputPath)
	cOutputPath := C.CString(outputPath)
	defer C.free(unsafe.Pointer(cInputPath))
	defer C.free(unsafe.Pointer(cOutputPath))

	// ✅ **Fix for `C.embed_qr_in_dct` missing**
	// Make sure the function is defined *before* importing `C`
	C.embed_qr_in_dct(cInputPath, cOutputPath, (*C.uchar)(unsafe.Pointer(&bitstream[0])), C.int(len(bitstream)))
	fmt.Println("Modified JPEG saved as:", outputPath)

	return nil
}

//go:build cgo
// +build cgo

package core

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -ljpeg
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <jpeglib.h>

// Forward declarations of the C functions we'll call
int embed_qr_in_dct_single(const char *input_path, const char *output_path, unsigned char *qr_data, int qr_size);
int embed_qr_in_dct_multi(const char *input_path, const char *output_path, unsigned char *qr_data, int qr_size);
void extract_qr_from_dct_single(const char *input_path, unsigned char *qr_data, int qr_size);
void extract_qr_from_dct_multi(const char *input_path, unsigned char *qr_data, int qr_size);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// CgoDCTProcessor implements DCTProcessor interface using CGO
type CgoDCTProcessor struct{}

func NewCgoDCTProcessor() *CgoDCTProcessor {
	return &CgoDCTProcessor{}
}

// EmbedData embeds data into DCT coefficients using CGO
func (p *CgoDCTProcessor) EmbedData(inputPath, outputPath string, data []byte, strategy DCTStrategy) error {
	// Convert paths to C strings
	cInputPath := C.CString(inputPath)
	cOutputPath := C.CString(outputPath)
	defer C.free(unsafe.Pointer(cInputPath))
	defer C.free(unsafe.Pointer(cOutputPath))

	var result C.int
	switch strategy {
	case DCTStrategyMulti:
		result = C.embed_qr_in_dct_multi(cInputPath, cOutputPath, (*C.uchar)(unsafe.Pointer(&data[0])), C.int(len(data)))
	default:
		result = C.embed_qr_in_dct_single(cInputPath, cOutputPath, (*C.uchar)(unsafe.Pointer(&data[0])), C.int(len(data)))
	}

	if result != 0 {
		return fmt.Errorf("DCT embedding failed with code %d (%s strategy)", int(result), strategy.String())
	}

	return nil
}

// ExtractData extracts data from DCT coefficients using CGO
func (p *CgoDCTProcessor) ExtractData(inputPath string, dataSize int, strategy DCTStrategy) ([]byte, error) {
	// Convert path to C string
	cInputPath := C.CString(inputPath)
	defer C.free(unsafe.Pointer(cInputPath))

	// Prepare buffer for extracted data
	extractedData := make([]byte, dataSize)

	// Call appropriate C function based on strategy
	switch strategy {
	case DCTStrategyMulti:
		C.extract_qr_from_dct_multi(cInputPath, (*C.uchar)(unsafe.Pointer(&extractedData[0])), C.int(dataSize))
	default:
		C.extract_qr_from_dct_single(cInputPath, (*C.uchar)(unsafe.Pointer(&extractedData[0])), C.int(dataSize))
	}

	return extractedData, nil
}

// CalculateCapacity calculates DCT capacity for given dimensions and strategy
func (p *CgoDCTProcessor) CalculateCapacity(width, height int, strategy DCTStrategy) int {
	dctBlocksX := (width + 7) / 8
	dctBlocksY := (height + 7) / 8
	return dctBlocksX * dctBlocksY * strategy.GetCoefficientsPerBit()
}

package core

import (
	"image"
)

// ImageProcessor handles image operations (SRP - Single Responsibility)
type ImageProcessor interface {
	GetDimensions(imagePath string) (*ImageDimensions, error)
	DecodeJPEG(imagePath string) (image.Image, error)
	EncodePNG(img image.Image, outputPath string) error
}

// QRCodeProcessor handles QR code operations (SRP)
type QRCodeProcessor interface {
	GenerateQR(data string, size int, eccLevel ECCLevel) ([]byte, error)
	ReadQR(imagePath string) (string, error)
	ConvertToBitstream(pngData []byte) ([]byte, error)
	ConvertFromBitstream(bitstream []byte, size int) (image.Image, error)
}

// DCTProcessor handles DCT coefficient operations (SRP)
type DCTProcessor interface {
	EmbedData(inputPath, outputPath string, data []byte, strategy DCTStrategy) error
	ExtractData(inputPath string, dataSize int, strategy DCTStrategy) ([]byte, error)
	CalculateCapacity(width, height int, strategy DCTStrategy) int
}

// QRSizeCalculator calculates optimal QR sizes (SRP)
type QRSizeCalculator interface {
	CalculateOptimalSize(imagePath string, payloadSize int, strategy DCTStrategy) (int, error)
	GetCapacityMap() map[int][2]int
}

// MetadataManager handles QR metadata storage/retrieval (SRP)
type MetadataManager interface {
	StoreQRMetadata(pixelSize, dataSize, dataArea int, env string) error
	RetrieveQRMetadata(env string) (pixelSize, dataSize, dataArea int, err error)
	StoreDCTStrategy(strategy DCTStrategy, env string) error
	RetrieveDCTStrategy(env string) (DCTStrategy, error)
}

// Concrete types
type ImageDimensions struct {
	Width  int
	Height int
}

type ECCLevel int

const (
	ECCLevelLow     ECCLevel = iota
	ECCLevelMedium  ECCLevel = iota
	ECCLevelHigh    ECCLevel = iota
	ECCLevelHighest ECCLevel = iota
)

type DCTStrategy int

const (
	DCTStrategySingle DCTStrategy = iota
	DCTStrategyMulti  DCTStrategy = iota
)

func (s DCTStrategy) String() string {
	switch s {
	case DCTStrategySingle:
		return "single-coefficient"
	case DCTStrategyMulti:
		return "multi-coefficient"
	default:
		return "unknown"
	}
}

func (s DCTStrategy) GetCoefficientsPerBit() int {
	switch s {
	case DCTStrategySingle:
		return 1
	case DCTStrategyMulti:
		return 2
	default:
		return 1
	}
}

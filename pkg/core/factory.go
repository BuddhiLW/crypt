package core

import (
	"fmt"
	"runtime"
)

// ServiceFactory creates configured services with all dependencies
type ServiceFactory struct{}

// NewServiceFactory creates a new factory
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// CreateSteganographyService creates a fully configured steganography service
func (f *ServiceFactory) CreateSteganographyService(env string) *SteganographyService {
	imageProcessor := NewJPEGImageProcessor()
	qrProcessor := NewGoQRProcessor()
	sizeCalculator := NewStandardQRSizeCalculatorWithProcessor(imageProcessor)
	metadataManager := NewBonzaiMetadataManager(env)

	// Use real DCT processor based on CGO availability
	var dctProcessor DCTProcessor
	if runtime.GOOS != "js" && runtime.GOARCH != "wasm" {
		// Use real CGO processor when available
		dctProcessor = NewCgoDCTProcessor()
		fmt.Printf("DEBUG: Using CGO DCT processor for %s environment\n", env)
	} else {
		// Fall back to mock only in environments where CGO is not available
		dctProcessor = NewMockDCTProcessor()
		fmt.Printf("DEBUG: Using mock DCT processor (CGO unavailable) for %s environment\n", env)
	}

	return NewSteganographyService(
		imageProcessor,
		qrProcessor,
		dctProcessor,
		sizeCalculator,
		metadataManager,
	)
}

// CreateTestSteganographyService creates a service with mock components for testing
func (f *ServiceFactory) CreateTestSteganographyService(env string) *SteganographyService {
	imageProcessor := NewMockJPEGImageProcessor()
	qrProcessor := NewGoQRProcessor()
	sizeCalculator := NewStandardQRSizeCalculatorWithProcessor(imageProcessor)
	metadataManager := NewBonzaiMetadataManager(env)

	// Use mock DCT processor for testing
	dctProcessor := NewMockDCTProcessor()

	return NewSteganographyService(
		imageProcessor,
		qrProcessor,
		dctProcessor,
		sizeCalculator,
		metadataManager,
	)
}

// CreateSteganographyServiceWithCGO creates a service with real CGO DCT processor
// This will be implemented when CGO is available
func (f *ServiceFactory) CreateSteganographyServiceWithCGO(env string) *SteganographyService {
	// For now, use mock processor
	return f.CreateSteganographyService(env)
}

// CgoDCTProcessor is implemented in cgo_dct_processor.go

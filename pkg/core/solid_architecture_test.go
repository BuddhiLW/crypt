package core

import (
	"testing"
)

// TestServiceFactory tests the factory pattern
func TestServiceFactory(t *testing.T) {
	factory := NewServiceFactory()
	if factory == nil {
		t.Fatal("Factory should not be nil")
	}

	service := factory.CreateTestSteganographyService("test-env")
	if service == nil {
		t.Fatal("Service should not be nil")
	}

	// Verify all dependencies are properly injected
	if service.imageProcessor == nil {
		t.Error("ImageProcessor should be injected")
	}
	if service.qrProcessor == nil {
		t.Error("QRProcessor should be injected")
	}
	if service.dctProcessor == nil {
		t.Error("DCTProcessor should be injected")
	}
	if service.sizeCalculator == nil {
		t.Error("SizeCalculator should be injected")
	}
	if service.metadataManager == nil {
		t.Error("MetadataManager should be injected")
	}
}

// TestDCTStrategy tests DCT strategy implementations
func TestDCTStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy DCTStrategy
		expected int
	}{
		{"Single", DCTStrategySingle, 1},
		{"Multi", DCTStrategyMulti, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coeffs := tt.strategy.GetCoefficientsPerBit()
			if coeffs != tt.expected {
				t.Errorf("Expected %d coefficients, got %d", tt.expected, coeffs)
			}
		})
	}
}

// TestQRSizeCalculator tests QR size calculation
func TestQRSizeCalculator(t *testing.T) {
	factory := NewServiceFactory()
	service := factory.CreateTestSteganographyService("test-env")
	calculator := service.sizeCalculator

	// Test basic functionality
	size, err := calculator.CalculateOptimalSize("test.jpg", 100, DCTStrategySingle)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if size <= 0 {
		t.Error("Size should be positive")
	}

	// Test capacity map
	capacityMap := calculator.GetCapacityMap()
	if len(capacityMap) == 0 {
		t.Error("Capacity map should not be empty")
	}

	// Test that we can get capacity for a known size
	if _, exists := capacityMap[256]; !exists {
		t.Error("Capacity map should contain standard QR sizes")
	}
}

// TestMetadataManager tests metadata management
func TestMetadataManager(t *testing.T) {
	manager := NewBonzaiMetadataManager("test-env")

	// Test storing and retrieving QR metadata
	err := manager.StoreQRMetadata(40, 1024, 1600, "test-env")
	if err != nil {
		t.Errorf("Failed to store QR metadata: %v", err)
	}

	pixelSize, dataSize, dataArea, err := manager.RetrieveQRMetadata("test-env")
	if err != nil {
		t.Errorf("Failed to retrieve QR metadata: %v", err)
	}

	if pixelSize != 40 || dataSize != 1024 || dataArea != 1600 {
		t.Errorf("QR metadata mismatch: expected (40, 1024, 1600), got (%d, %d, %d)", pixelSize, dataSize, dataArea)
	}

	// Test storing and retrieving DCT strategy
	err = manager.StoreDCTStrategy(DCTStrategyMulti, "test-env")
	if err != nil {
		t.Errorf("Failed to store DCT strategy: %v", err)
	}

	strategy, err := manager.RetrieveDCTStrategy("test-env")
	if err != nil {
		t.Errorf("Failed to retrieve DCT strategy: %v", err)
	}

	if strategy != DCTStrategyMulti {
		t.Errorf("DCT strategy mismatch: expected %v, got %v", DCTStrategyMulti, strategy)
	}
}

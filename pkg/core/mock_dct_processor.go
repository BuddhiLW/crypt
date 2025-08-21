package core

import (
	"fmt"
	"os"
)

// MockDCTProcessor implements DCTProcessor interface for testing
type MockDCTProcessor struct{}

func NewMockDCTProcessor() *MockDCTProcessor {
	return &MockDCTProcessor{}
}

// EmbedData mocks embedding data into DCT coefficients
func (p *MockDCTProcessor) EmbedData(inputPath, outputPath string, data []byte, strategy DCTStrategy) error {
	// Mock implementation - just validate inputs
	if inputPath == "" {
		return fmt.Errorf("input path cannot be empty")
	}
	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	if len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}

	// For testing, actually copy the input file to the output file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Copy the input file to output (simulating embedding)
	_, err = outputFile.ReadFrom(inputFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// ExtractData mocks extracting data from DCT coefficients
func (p *MockDCTProcessor) ExtractData(inputPath string, dataSize int, strategy DCTStrategy) ([]byte, error) {
	// Mock implementation - return dummy data
	if inputPath == "" {
		return nil, fmt.Errorf("input path cannot be empty")
	}
	if dataSize <= 0 {
		return nil, fmt.Errorf("data size must be positive")
	}

	// Return mock data
	mockData := make([]byte, dataSize)
	for i := range mockData {
		mockData[i] = byte(i % 256)
	}
	return mockData, nil
}

// CalculateCapacity calculates DCT capacity for given dimensions and strategy
func (p *MockDCTProcessor) CalculateCapacity(width, height int, strategy DCTStrategy) int {
	dctBlocksX := (width + 7) / 8
	dctBlocksY := (height + 7) / 8
	return dctBlocksX * dctBlocksY * strategy.GetCoefficientsPerBit()
}

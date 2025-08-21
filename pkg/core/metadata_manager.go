package core

import (
	"fmt"
	"strconv"

	"github.com/rwxrob/bonzai/vars"
)

const (
	QRSizeVar      = "qr-size"
	QRDataSizeVar  = "qr-data-size"
	QRDataAreaVar  = "qr-data-area"
	DCTStrategyVar = "dct-strategy"
)

// BonzaiMetadataManager implements MetadataManager using Bonzai vars
type BonzaiMetadataManager struct {
	env string
}

func NewBonzaiMetadataManager(env string) *BonzaiMetadataManager {
	return &BonzaiMetadataManager{env: env}
}

func (m *BonzaiMetadataManager) StoreQRMetadata(pixelSize, dataSize, dataArea int, env string) error {
	// Store pixel size
	if err := vars.Set(QRSizeVar, fmt.Sprintf("%d", pixelSize), env); err != nil {
		return fmt.Errorf("failed to store QR pixel size: %w", err)
	}

	// Store data size
	if err := vars.Set(QRDataSizeVar, fmt.Sprintf("%d", dataSize), env); err != nil {
		return fmt.Errorf("failed to store QR data size: %w", err)
	}

	// Store data area
	if err := vars.Set(QRDataAreaVar, fmt.Sprintf("%d", dataArea), env); err != nil {
		return fmt.Errorf("failed to store QR data area: %w", err)
	}

	return nil
}

func (m *BonzaiMetadataManager) RetrieveQRMetadata(env string) (pixelSize, dataSize, dataArea int, err error) {
	// Retrieve pixel size
	pixelSizeStr, _ := vars.Get(QRSizeVar, env)
	if pixelSizeStr == "" {
		pixelSize = 256 // Default size
	} else {
		pixelSize, err = strconv.Atoi(pixelSizeStr)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid QR pixel size in vars: %w", err)
		}
	}

	// Retrieve data size
	dataSizeStr, _ := vars.Get(QRDataSizeVar, env)
	if dataSizeStr != "" {
		dataSize, err = strconv.Atoi(dataSizeStr)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid QR data size in vars: %w", err)
		}
	}

	// Retrieve data area
	dataAreaStr, _ := vars.Get(QRDataAreaVar, env)
	if dataAreaStr != "" {
		dataArea, err = strconv.Atoi(dataAreaStr)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid QR data area in vars: %w", err)
		}
	} else {
		dataArea = pixelSize // Fallback to pixel size
	}

	return pixelSize, dataSize, dataArea, nil
}

func (m *BonzaiMetadataManager) StoreDCTStrategy(strategy DCTStrategy, env string) error {
	strategyName := strategy.String()
	if err := vars.Set(DCTStrategyVar, strategyName, env); err != nil {
		return fmt.Errorf("failed to store DCT strategy: %w", err)
	}
	return nil
}

func (m *BonzaiMetadataManager) RetrieveDCTStrategy(env string) (DCTStrategy, error) {
	strategyName, _ := vars.Get(DCTStrategyVar, env)
	if strategyName == "" {
		return DCTStrategySingle, nil // Default strategy
	}

	switch strategyName {
	case "single-coefficient":
		return DCTStrategySingle, nil
	case "multi-coefficient":
		return DCTStrategyMulti, nil
	default:
		return DCTStrategySingle, nil // Default fallback
	}
}

package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// MultiQRMetadata represents the metadata for multi-QR grid embedding
type MultiQRMetadata struct {
	Version       string               `json:"version"`
	TotalChunks   int                  `json:"total_chunks"`
	GridSize      [2]int               `json:"grid_size"`       // [rows, cols]
	ChunkSize     int                  `json:"chunk_size"`      // bytes per chunk
	TotalDataSize int                  `json:"total_data_size"` // total payload size
	ChunkHashes   map[string]ChunkInfo `json:"chunk_hashes"`    // hash -> chunk info
	HashOrder     []string             `json:"hash_order"`      // ordered list of hashes
	ECCLevel      string               `json:"ecc_level"`       // "H" for High ECC
	Compression   string               `json:"compression"`     // "none" or compression type
	Checksum      string               `json:"checksum"`        // SHA256 of original data
	Timestamp     int64                `json:"timestamp"`
}

// ChunkInfo contains information about a single chunk
type ChunkInfo struct {
	Index    int    `json:"index"`     // 0-based chunk index
	Position [2]int `json:"position"`  // [row, col] in grid
	Size     int    `json:"size"`      // actual chunk size in bytes
	Hash     string `json:"hash"`      // SHA256 of chunk data
	FileName string `json:"file_name"` // suggested filename (optional)
}

// NewMultiQRMetadata creates a new metadata structure
func NewMultiQRMetadata(totalDataSize, chunkSize int, gridSize [2]int) *MultiQRMetadata {
	totalChunks := (totalDataSize + chunkSize - 1) / chunkSize // ceiling division

	return &MultiQRMetadata{
		Version:       "1.0",
		TotalChunks:   totalChunks,
		GridSize:      gridSize,
		ChunkSize:     chunkSize,
		TotalDataSize: totalDataSize,
		ChunkHashes:   make(map[string]ChunkInfo),
		HashOrder:     make([]string, 0, totalChunks),
		ECCLevel:      "H", // High ECC for compression resilience
		Compression:   "none",
		Timestamp:     time.Now().Unix(),
	}
}

// AddChunk adds a chunk to the metadata
func (m *MultiQRMetadata) AddChunk(index int, data []byte, position [2]int, fileName string) error {
	if index >= m.TotalChunks {
		return fmt.Errorf("chunk index %d exceeds total chunks %d", index, m.TotalChunks)
	}

	// Calculate hash of chunk data
	hash := calculateChunkHash(data)

	chunkInfo := ChunkInfo{
		Index:    index,
		Position: position,
		Size:     len(data),
		Hash:     hash,
		FileName: fileName,
	}

	m.ChunkHashes[hash] = chunkInfo
	m.HashOrder = append(m.HashOrder, hash)

	return nil
}

// GetChunkByHash retrieves chunk info by hash
func (m *MultiQRMetadata) GetChunkByHash(hash string) (ChunkInfo, bool) {
	info, exists := m.ChunkHashes[hash]
	return info, exists
}

// GetChunkByIndex retrieves chunk info by index
func (m *MultiQRMetadata) GetChunkByIndex(index int) (ChunkInfo, bool) {
	if index >= len(m.HashOrder) {
		return ChunkInfo{}, false
	}

	hash := m.HashOrder[index]
	return m.GetChunkByHash(hash)
}

// GetOrderedChunks returns chunks in the correct order
func (m *MultiQRMetadata) GetOrderedChunks() []ChunkInfo {
	chunks := make([]ChunkInfo, len(m.HashOrder))
	for i, hash := range m.HashOrder {
		chunks[i] = m.ChunkHashes[hash]
	}
	return chunks
}

// ValidateChunk validates a chunk against metadata
func (m *MultiQRMetadata) ValidateChunk(data []byte, hash string) error {
	expectedHash := calculateChunkHash(data)
	if expectedHash != hash {
		return fmt.Errorf("chunk hash mismatch: expected %s, got %s", expectedHash, expectedHash)
	}

	chunkInfo, exists := m.GetChunkByHash(hash)
	if !exists {
		return fmt.Errorf("chunk hash %s not found in metadata", hash)
	}

	if len(data) != chunkInfo.Size {
		return fmt.Errorf("chunk size mismatch: expected %d, got %d", chunkInfo.Size, len(data))
	}

	return nil
}

// ToJSON converts metadata to JSON
func (m *MultiQRMetadata) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// FromJSON creates metadata from JSON
func FromJSON(data []byte) (*MultiQRMetadata, error) {
	var metadata MultiQRMetadata
	err := json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}
	return &metadata, nil
}

// calculateChunkHash calculates SHA256 hash of chunk data
func calculateChunkHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// MultiQRFileScanner scans directories for QR chunk files
type MultiQRFileScanner struct{}

// NewMultiQRFileScanner creates a new file scanner
func NewMultiQRFileScanner() *MultiQRFileScanner {
	return &MultiQRFileScanner{}
}

// ScanDirectory scans a directory for QR chunk files and metadata
func (s *MultiQRFileScanner) ScanDirectory(dirPath string) (*MultiQRMetadata, []string, error) {
	// This would be implemented to scan for QR files
	// For now, return a mock implementation
	return nil, nil, fmt.Errorf("directory scanning not yet implemented")
}

// FindMetadataFile finds the metadata QR file in a directory
func (s *MultiQRFileScanner) FindMetadataFile(dirPath string) (string, error) {
	// This would scan for files containing metadata
	// For now, return a mock implementation
	return "", fmt.Errorf("metadata file scanning not yet implemented")
}

// ValidateChunkFiles validates all chunk files against metadata
func (s *MultiQRFileScanner) ValidateChunkFiles(metadata *MultiQRMetadata, chunkFiles []string) error {
	if len(chunkFiles) != metadata.TotalChunks {
		return fmt.Errorf("expected %d chunk files, found %d", metadata.TotalChunks, len(chunkFiles))
	}

	// Validate each chunk file
	for range chunkFiles {
		// Read file and validate against metadata
		// This would be implemented to actually read and validate files
	}

	return nil
}

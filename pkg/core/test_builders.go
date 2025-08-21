package core

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// MetadataBuilder creates test MultiQRMetadata instances
type MetadataBuilder struct {
	metadata *MultiQRMetadata
}

// NewMetadataBuilder creates a new metadata builder with sensible defaults
func NewMetadataBuilder() *MetadataBuilder {
	return &MetadataBuilder{
		metadata: &MultiQRMetadata{
			Version:     "1.0",
			TotalChunks: 4,
			GridSize:    [2]int{2, 2},
			ChunkSize:   256,
			ECCLevel:    "H",
			Compression: "none",
			Timestamp:   time.Now().Unix(),
		},
	}
}

// WithChunks sets the number of chunks
func (b *MetadataBuilder) WithChunks(count int) *MetadataBuilder {
	b.metadata.TotalChunks = count
	return b
}

// WithGridSize sets the grid dimensions
func (b *MetadataBuilder) WithGridSize(rows, cols int) *MetadataBuilder {
	b.metadata.GridSize = [2]int{rows, cols}
	return b
}

// WithChunkSize sets the chunk size in bytes
func (b *MetadataBuilder) WithChunkSize(size int) *MetadataBuilder {
	b.metadata.ChunkSize = size
	return b
}

// WithTotalDataSize sets the total data size
func (b *MetadataBuilder) WithTotalDataSize(size int) *MetadataBuilder {
	b.metadata.TotalDataSize = size
	return b
}

// WithECCLevel sets the ECC level
func (b *MetadataBuilder) WithECCLevel(level string) *MetadataBuilder {
	b.metadata.ECCLevel = level
	return b
}

// WithChecksum sets the data checksum
func (b *MetadataBuilder) WithChecksum(data []byte) *MetadataBuilder {
	hash := sha256.Sum256(data)
	b.metadata.Checksum = fmt.Sprintf("%x", hash)
	return b
}

// Build creates the final MultiQRMetadata
func (b *MetadataBuilder) Build() *MultiQRMetadata {
	return b.metadata
}

// ChunkInfoBuilder creates test ChunkInfo instances
type ChunkInfoBuilder struct {
	chunkInfo ChunkInfo
}

// NewChunkInfoBuilder creates a new chunk info builder
func NewChunkInfoBuilder() *ChunkInfoBuilder {
	return &ChunkInfoBuilder{
		chunkInfo: ChunkInfo{
			Index:    0,
			Position: [2]int{0, 0},
			Size:     256,
			Hash:     "",
			FileName: "chunk_0.qr",
		},
	}
}

// WithIndex sets the chunk index
func (b *ChunkInfoBuilder) WithIndex(index int) *ChunkInfoBuilder {
	b.chunkInfo.Index = index
	return b
}

// WithPosition sets the grid position
func (b *ChunkInfoBuilder) WithPosition(row, col int) *ChunkInfoBuilder {
	b.chunkInfo.Position = [2]int{row, col}
	return b
}

// WithSize sets the chunk size
func (b *ChunkInfoBuilder) WithSize(size int) *ChunkInfoBuilder {
	b.chunkInfo.Size = size
	return b
}

// WithData sets the chunk data and calculates hash
func (b *ChunkInfoBuilder) WithData(data []byte) *ChunkInfoBuilder {
	b.chunkInfo.Size = len(data)
	b.chunkInfo.Hash = calculateChunkHash(data)
	return b
}

// WithFileName sets the suggested filename
func (b *ChunkInfoBuilder) WithFileName(name string) *ChunkInfoBuilder {
	b.chunkInfo.FileName = name
	return b
}

// Build creates the final ChunkInfo
func (b *ChunkInfoBuilder) Build() ChunkInfo {
	return b.chunkInfo
}

// TestDataBuilder creates test data for various scenarios
type TestDataBuilder struct{}

// NewTestDataBuilder creates a new test data builder
func NewTestDataBuilder() *TestDataBuilder {
	return &TestDataBuilder{}
}

// SmallPayload creates a small test payload
func (b *TestDataBuilder) SmallPayload() []byte {
	return []byte("small test payload")
}

// MediumPayload creates a medium test payload
func (b *TestDataBuilder) MediumPayload() []byte {
	return []byte("medium test payload with more data to test chunking functionality")
}

// LargePayload creates a large test payload
func (b *TestDataBuilder) LargePayload() []byte {
	// Create a 1KB payload
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	return payload
}

// ChunkedPayload creates a payload that will be split into chunks
func (b *TestDataBuilder) ChunkedPayload(chunkSize int, numChunks int) [][]byte {
	totalSize := chunkSize * numChunks
	payload := make([]byte, totalSize)

	// Fill with test data
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	// Split into chunks
	chunks := make([][]byte, numChunks)
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		chunks[i] = payload[start:end]
	}

	return chunks
}

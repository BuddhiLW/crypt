package core_test

import (
	"fmt"
	"testing"

	"github.com/BuddhiLW/crypt/pkg/core"
)

func TestMultiQRMetadataCreation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		dataSize       int
		chunkSize      int
		gridSize       [2]int
		expectedChunks int
	}{
		{"small payload", 512, 256, [2]int{2, 2}, 2},
		{"medium payload", 1024, 256, [2]int{2, 2}, 4},
		{"large payload", 2048, 256, [2]int{3, 3}, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := core.NewMultiQRMetadata(tt.dataSize, tt.chunkSize, tt.gridSize)

			if metadata.TotalChunks != tt.expectedChunks {
				t.Errorf("expected %d chunks, got %d", tt.expectedChunks, metadata.TotalChunks)
			}

			if metadata.ChunkSize != tt.chunkSize {
				t.Errorf("expected chunk size %d, got %d", tt.chunkSize, metadata.ChunkSize)
			}

			if metadata.GridSize != tt.gridSize {
				t.Errorf("expected grid size %v, got %v", tt.gridSize, metadata.GridSize)
			}

			if metadata.ECCLevel != "H" {
				t.Errorf("expected ECC level H, got %s", metadata.ECCLevel)
			}
		})
	}
}

func TestMultiQRMetadataWithBuilder(t *testing.T) {
	t.Parallel()

	metadata := core.NewMetadataBuilder().
		WithChunks(6).
		WithGridSize(2, 3).
		WithChunkSize(512).
		WithTotalDataSize(2048).
		WithECCLevel("H").
		Build()

	if metadata.TotalChunks != 6 {
		t.Errorf("expected 6 chunks, got %d", metadata.TotalChunks)
	}

	if metadata.GridSize != [2]int{2, 3} {
		t.Errorf("expected grid size [2,3], got %v", metadata.GridSize)
	}

	if metadata.ChunkSize != 512 {
		t.Errorf("expected chunk size 512, got %d", metadata.ChunkSize)
	}
}

func TestChunkManagement(t *testing.T) {
	t.Parallel()

	metadata := core.NewMultiQRMetadata(1024, 256, [2]int{2, 2})
	testDataBuilder := core.NewTestDataBuilder()

	// Add test chunks
	testChunks := testDataBuilder.ChunkedPayload(256, 4)
	for i, chunkData := range testChunks {
		row := i / 2
		col := i % 2
		err := metadata.AddChunk(i, chunkData, [2]int{row, col}, fmt.Sprintf("chunk_%d.qr", i))
		if err != nil {
			t.Fatalf("failed to add chunk %d: %v", i, err)
		}
	}

	// Verify chunks were added
	if len(metadata.HashOrder) != 4 {
		t.Errorf("expected 4 chunks in hash order, got %d", len(metadata.HashOrder))
	}

	// Test chunk retrieval by hash order
	for i, hash := range metadata.HashOrder {
		chunkInfo, exists := metadata.GetChunkByHash(hash)
		if !exists {
			t.Errorf("chunk with hash %s not found", hash)
			continue
		}

		// The index should match the order in which chunks were added
		if chunkInfo.Index != i {
			t.Errorf("chunk %d has wrong index %d", i, chunkInfo.Index)
		}

		// The position should match the grid layout
		expectedRow := i / 2
		expectedCol := i % 2
		if chunkInfo.Position != [2]int{expectedRow, expectedCol} {
			t.Errorf("chunk %d has wrong position %v, expected [%d,%d]",
				i, chunkInfo.Position, expectedRow, expectedCol)
		}
	}
}

func TestHashBasedIdentification(t *testing.T) {
	t.Parallel()

	metadata := core.NewMultiQRMetadata(512, 256, [2]int{1, 2})
	testData := []byte("test_chunk_data_for_hashing")

	// Add chunk with builder
	chunkInfo := core.NewChunkInfoBuilder().
		WithIndex(0).
		WithPosition(0, 0).
		WithData(testData).
		WithFileName("test_chunk.qr").
		Build()

	err := metadata.AddChunk(chunkInfo.Index, testData, chunkInfo.Position, chunkInfo.FileName)
	if err != nil {
		t.Fatalf("failed to add test chunk: %v", err)
	}

	// Test validation with correct data
	err = metadata.ValidateChunk(testData, chunkInfo.Hash)
	if err != nil {
		t.Errorf("chunk validation failed: %v", err)
	}

	// Test validation with wrong data
	wrongData := []byte("wrong_chunk_data")
	err = metadata.ValidateChunk(wrongData, chunkInfo.Hash)
	if err == nil {
		t.Error("validation should have failed with wrong data")
	}
}

func TestJSONSerialization(t *testing.T) {
	t.Parallel()

	metadata := core.NewMultiQRMetadata(200, 128, [2]int{1, 2})

	// Add some test data
	testData := []byte("test data")
	err := metadata.AddChunk(0, testData, [2]int{0, 0}, "chunk_0.qr")
	if err != nil {
		t.Fatalf("failed to add chunk: %v", err)
	}

	// Test serialization
	jsonData, err := metadata.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize metadata: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON data is empty")
	}

	// Test deserialization
	parsedMetadata, err := core.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}

	if parsedMetadata.TotalChunks != metadata.TotalChunks {
		t.Errorf("parsed metadata has wrong chunk count: expected %d, got %d",
			metadata.TotalChunks, parsedMetadata.TotalChunks)
	}

	if parsedMetadata.GridSize != metadata.GridSize {
		t.Errorf("parsed metadata has wrong grid size: expected %v, got %v",
			metadata.GridSize, parsedMetadata.GridSize)
	}
}

func TestErrorConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *core.MultiQRMetadata
		test    func(*core.MultiQRMetadata) error
		wantErr bool
	}{
		{
			name: "chunk index out of bounds",
			setup: func() *core.MultiQRMetadata {
				return core.NewMultiQRMetadata(512, 256, [2]int{1, 1})
			},
			test: func(m *core.MultiQRMetadata) error {
				return m.AddChunk(5, []byte("data"), [2]int{0, 0}, "test.qr")
			},
			wantErr: true,
		},
		{
			name: "invalid chunk hash",
			setup: func() *core.MultiQRMetadata {
				m := core.NewMultiQRMetadata(512, 256, [2]int{1, 1})
				m.AddChunk(0, []byte("data"), [2]int{0, 0}, "test.qr")
				return m
			},
			test: func(m *core.MultiQRMetadata) error {
				return m.ValidateChunk([]byte("wrong data"), "invalid_hash")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := tt.setup()
			err := tt.test(metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

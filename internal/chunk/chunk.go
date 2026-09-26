package chunk

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"synx/internal/integrity"
)

const (
	DefaultChunkSize int64 = 8 * 1024 * 1024 // 8 MB
)

type Chunk struct {
	TransferID string `json:"transfer_id"`
	Index      int    `json:"chunk_index"`
	Offset     int64  `json:"offset"`
	Size       int64  `json:"size"`
	SHA256     string `json:"sha256"`
	Data       []byte `json:"-"`
}

type Manifest struct {
	mu         sync.RWMutex
	TransferID string        `json:"transfer_id"`
	FileName   string        `json:"file_name"`
	FileSize   int64         `json:"file_size"`
	ChunkSize  int64         `json:"chunk_size"`
	Total      int           `json:"total_chunks"`
	Completed  map[int]int64 `json:"completed"` // index -> bytes written
}

func NewManifest(transferID, fileName string, fileSize, chunkSize int64) *Manifest {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	total := int((fileSize + chunkSize - 1) / chunkSize)
	if total == 0 {
		total = 1
	}

	return &Manifest{
		TransferID: transferID,
		FileName:   fileName,
		FileSize:   fileSize,
		ChunkSize:  chunkSize,
		Total:      total,
		Completed:  make(map[int]int64),
	}
}

func (m *Manifest) MarkCompleted(index int, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Completed[index] = size
}

func (m *Manifest) IsCompleted(index int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.Completed[index]
	return ok
}

func (m *Manifest) CompletedBytes() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var sum int64
	for _, sz := range m.Completed {
		sum += sz
	}
	return sum
}

func (m *Manifest) CompletedIndices() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	indices := make([]int, 0, len(m.Completed))
	for idx := range m.Completed {
		indices = append(indices, idx)
	}
	return indices
}

// ReadChunk reads a chunk of bytes from source file at chunk index
func ReadChunk(filePath string, index int, chunkSize int64) (*Chunk, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open source file for chunk read: %w", err)
	}
	defer f.Close()

	offset := int64(index) * chunkSize
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek error at chunk %d: %w", index, err)
	}

	buf := make([]byte, chunkSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("read error at chunk %d: %w", index, err)
	}

	data := buf[:n]
	hash := integrity.BytesSHA256(data)

	return &Chunk{
		Index:  index,
		Offset: offset,
		Size:   int64(n),
		SHA256: hash,
		Data:   data,
	}, nil
}

// WriteChunk writes chunk data to destination part file at specified offset
func WriteChunk(partPath string, c *Chunk, expectedHash string) error {
	if expectedHash != "" && c.SHA256 != "" && c.SHA256 != expectedHash {
		return fmt.Errorf("chunk %d checksum mismatch: expected %s, got %s", c.Index, expectedHash, c.SHA256)
	}

	// Verify data hash matches claimed hash
	computed := integrity.BytesSHA256(c.Data)
	if c.SHA256 != "" && computed != c.SHA256 {
		return fmt.Errorf("chunk %d payload corrupted: hash mismatch", c.Index)
	}

	if err := os.MkdirAll(filepath.Dir(partPath), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open part file for writing: %w", err)
	}
	defer f.Close()

	if _, err := f.Seek(c.Offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek failed for chunk %d: %w", c.Index, err)
	}

	if _, err := f.Write(c.Data); err != nil {
		return fmt.Errorf("failed to write chunk %d: %w", c.Index, err)
	}

	return nil
}

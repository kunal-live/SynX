package protocol

const (
	CurrentVersion = 1

	HeaderProtocolVersion = "X-SynX-Protocol-Version"
	HeaderToken           = "X-SynX-Token"
	HeaderSession         = "X-SynX-Session"
	HeaderDeviceID        = "X-SynX-Device-ID"
	HeaderChunkIndex      = "X-Chunk-Index"
	HeaderChunkOffset     = "X-Chunk-Offset"
	HeaderChunkHash       = "X-Chunk-Hash"
)

type Capabilities struct {
	ProtocolVersion int      `json:"protocol_version"`
	Transports      []string `json:"transports"`
	MaxChunkSize    int64    `json:"max_chunk_size"`
	Compression     bool     `json:"compression"`
	Resume          bool     `json:"resume"`
	ParallelChunks  bool     `json:"parallel_chunks"`
	FolderTransfer  bool     `json:"folder_transfer"`
}

func DefaultCapabilities() Capabilities {
	return Capabilities{
		ProtocolVersion: CurrentVersion,
		Transports:      []string{"http"},
		MaxChunkSize:    16 * 1024 * 1024,
		Compression:     false,
		Resume:          true,
		ParallelChunks:  true,
		FolderTransfer:  true,
	}
}

type TransferNegotiationRequest struct {
	SourcePath  string `json:"source"`
	Destination string `json:"destination"`
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
	SHA256      string `json:"sha256"`
	ChunkSize   int64  `json:"chunk_size"`
}

type TransferNegotiationResponse struct {
	TransferID   string `json:"transfer_id"`
	Accepted     bool   `json:"accepted"`
	ChunkSize    int64  `json:"chunk_size"`
	ResumeOffset int64  `json:"resume_offset"`
	Completed    []int  `json:"completed_chunks,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type CommitRequest struct {
	SHA256 string `json:"sha256"`
}

type CommitResponse struct {
	Status   string `json:"status"` // "completed" or "failed"
	Verified bool   `json:"verified"`
	Error    string `json:"error,omitempty"`
}

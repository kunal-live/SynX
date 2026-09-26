package history

import (
	"synx/internal/storage"
)

type Entry struct {
	ID          string `json:"id"`
	TransferID  string `json:"transfer_id"`
	PeerID      string `json:"peer_id"`
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
	Direction   string `json:"direction"`
	Status      string `json:"status"`
	CompletedAt int64  `json:"completed_at"`
}

type Service struct {
	db *storage.DB
}

func NewService(db *storage.DB) *Service {
	return &Service{db: db}
}

func (s *Service) List(limit int) ([]Entry, error) {
	if s.db == nil {
		return []Entry{}, nil
	}
	records, err := s.db.ListHistory(limit)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(records))
	for _, r := range records {
		out = append(out, Entry{
			ID:          r.ID,
			TransferID:  r.TransferID,
			PeerID:      r.PeerID,
			FileName:    r.FileName,
			Size:        r.Size,
			Direction:   r.Direction,
			Status:      r.Status,
			CompletedAt: r.CompletedAt,
		})
	}
	return out, nil
}

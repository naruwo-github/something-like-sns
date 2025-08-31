package port

import "time"

// CursorEncoder defines an interface for encoding and decoding cursors.
type CursorEncoder interface {
	Encode(t time.Time, id uint64) string
	Decode(token string) (time.Time, uint64, error)
}

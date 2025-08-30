package mysql

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type cursorEncoder struct{}

func NewCursorEncoder() *cursorEncoder {
	return &cursorEncoder{}
}

func (e *cursorEncoder) Decode(token string) (time.Time, uint64, error) {
	if token == "" {
		return time.Time{}, 0, nil
	}
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, 0, err
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return time.Time{}, 0, errors.New("bad cursor")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, 0, err
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	return t, id, err
}

func (e *cursorEncoder) Encode(t time.Time, id uint64) string {
	if t.IsZero() {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%d", t.Format(time.RFC3339Nano), id)))
}

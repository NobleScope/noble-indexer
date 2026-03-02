package helpers

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/pkg/errors"
)

var errInvalidCursorLength = errors.New("invalid cursor length")

// EncodeTimeIDCursor encodes a (time, id) pair into a base64url string.
// Binary format: 8 bytes UnixMicro (big-endian) + 8 bytes id (big-endian).
func EncodeTimeIDCursor(t time.Time, id uint64) string {
	var buf [16]byte
	binary.BigEndian.PutUint64(buf[:8], uint64(t.UnixMicro()))
	binary.BigEndian.PutUint64(buf[8:], id)
	return base64.RawURLEncoding.EncodeToString(buf[:])
}

// DecodeTimeIDCursor decodes a base64url cursor string into a (time, id) pair.
func DecodeTimeIDCursor(cursor string) (time.Time, uint64, error) {
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, 0, errors.Wrap(err, "decoding cursor")
	}
	if len(data) != 16 {
		return time.Time{}, 0, errInvalidCursorLength
	}
	t := time.UnixMicro(int64(binary.BigEndian.Uint64(data[:8]))).UTC()
	id := binary.BigEndian.Uint64(data[8:])
	return t, id, nil
}

// EncodeIDCursor encodes an id into a base64url string.
// Binary format: 8 bytes id (big-endian).
func EncodeIDCursor(id uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], id)
	return base64.RawURLEncoding.EncodeToString(buf[:])
}

// DecodeIDCursor decodes a base64url cursor string into an id.
func DecodeIDCursor(cursor string) (uint64, error) {
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, errors.Wrap(err, "decoding cursor")
	}
	if len(data) != 8 {
		return 0, errInvalidCursorLength
	}
	return binary.BigEndian.Uint64(data), nil
}

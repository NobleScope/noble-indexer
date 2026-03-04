package helpers

import (
	"testing"
	"time"
)

func TestEncodeDecodeTimeIDCursor(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		id   uint64
	}{
		{"zero time and id", time.Unix(0, 0).UTC(), 0},
		{"typical values", time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC), 12345},
		{"max id", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), ^uint64(0)},
		{"microsecond precision", time.Date(2024, 3, 1, 10, 20, 30, 123456000, time.UTC), 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := EncodeTimeIDCursor(tt.time, tt.id)
			gotTime, gotID, err := DecodeTimeIDCursor(cursor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Compare at microsecond precision (nanoseconds are lost)
			if gotTime.UnixMicro() != tt.time.UnixMicro() {
				t.Errorf("time mismatch: got %v, want %v", gotTime, tt.time)
			}
			if gotID != tt.id {
				t.Errorf("id mismatch: got %d, want %d", gotID, tt.id)
			}
		})
	}
}

func TestEncodeDecodeIDCursor(t *testing.T) {
	tests := []struct {
		name string
		id   uint64
	}{
		{"zero", 0},
		{"typical", 12345},
		{"max", ^uint64(0)},
		{"one", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := EncodeIDCursor(tt.id)
			gotID, err := DecodeIDCursor(cursor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotID != tt.id {
				t.Errorf("id mismatch: got %d, want %d", gotID, tt.id)
			}
		})
	}
}

func TestDecodeInvalidCursor(t *testing.T) {
	t.Run("invalid base64", func(t *testing.T) {
		_, _, err := DecodeTimeIDCursor("not-valid-base64!!!")
		if err == nil {
			t.Error("expected error for invalid base64")
		}
	})

	t.Run("wrong length for TimeID", func(t *testing.T) {
		_, _, err := DecodeTimeIDCursor(EncodeIDCursor(1)) // 8 bytes instead of 16
		if err == nil {
			t.Error("expected error for wrong length")
		}
	})

	t.Run("wrong length for ID", func(t *testing.T) {
		_, err := DecodeIDCursor(EncodeTimeIDCursor(time.Now(), 1)) // 16 bytes instead of 8
		if err == nil {
			t.Error("expected error for wrong length")
		}
	})

	t.Run("empty string TimeID", func(t *testing.T) {
		_, _, err := DecodeTimeIDCursor("")
		if err == nil {
			t.Error("expected error for empty string")
		}
	})

	t.Run("empty string ID", func(t *testing.T) {
		_, err := DecodeIDCursor("")
		if err == nil {
			t.Error("expected error for empty string")
		}
	})
}

func BenchmarkEncodeTimeIDCursor(b *testing.B) {
	t := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	id := uint64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeTimeIDCursor(t, id)
	}
}

func BenchmarkDecodeTimeIDCursor(b *testing.B) {
	cursor := EncodeTimeIDCursor(time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC), 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = DecodeTimeIDCursor(cursor)
	}
}

func BenchmarkEncodeIDCursor(b *testing.B) {
	id := uint64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeIDCursor(id)
	}
}

func BenchmarkDecodeIDCursor(b *testing.B) {
	cursor := EncodeIDCursor(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeIDCursor(cursor)
	}
}

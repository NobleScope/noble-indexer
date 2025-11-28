package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type Hex []byte

var nullBytes = "null"

func HexFromString(hexStr string) (Hex, error) {
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}

	resultBytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		b, err := strconv.ParseUint(hexStr[i:i+2], 16, 8)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid hex string")
		}
		resultBytes[i/2] = byte(b)
	}

	return resultBytes, nil
}

func (h *Hex) UnmarshalJSON(data []byte) error {
	if h == nil {
		return nil
	}

	if nullBytes == string(data) {
		*h = nil
		return nil
	}
	length := len(data)
	if data[0] != '"' || data[length-1] != '"' {
		return errors.Errorf("hex should be quotted string: got=%s", data)
	}

	data = bytes.Trim(data, `"`)
	data = bytes.TrimPrefix(data, []byte("0x"))

	if len(data)%2 == 1 {
		data = append([]byte{'0'}, data...)
	}

	*h = make(Hex, hex.DecodedLen(len(data)))
	if len(data) == 0 {
		return nil
	}
	_, err := hex.Decode(*h, data)
	return err
}

func (h Hex) MarshalJSON() ([]byte, error) {
	if len(h) == 0 {
		return []byte(nullBytes), nil
	}
	hexStr := hex.EncodeToString(h)
	if len(hexStr) > 0 && hexStr[0] == '0' {
		hexStr = hexStr[1:]
	}

	return []byte(strconv.Quote("0x" + hexStr)), nil
}

func (h *Hex) Scan(src interface{}) (err error) {
	switch val := src.(type) {
	case []byte:
		*h = make(Hex, len(val))
		_ = copy(*h, val)
	case nil:
		*h = make(Hex, 0)
	default:
		return errors.Errorf("unknown hex database type: %T", src)
	}
	return nil
}

var _ driver.Valuer = (*Hex)(nil)

func (h Hex) Value() (driver.Value, error) {
	return []byte(h), nil
}

func (h Hex) Bytes() []byte {
	return []byte(h)
}

func (h Hex) Hex() string {
	if len(h) == 0 {
		return ""
	}
	return "0x" + hex.EncodeToString(h)
}

func (h Hex) String() string {
	if len(h) == 0 {
		return ""
	}
	return hex.EncodeToString(h)
}

func (h Hex) Uint64() (uint64, error) {
	if len(h) == 0 {
		return 0, nil
	}

	b := trimToNBytes(h, 8)
	return binary.BigEndian.Uint64(b), nil
}

func (h Hex) Uint32() (uint32, error) {
	if len(h) == 0 {
		return 0, nil
	}

	b := trimToNBytes(h, 4)
	return binary.BigEndian.Uint32(b), nil
}

func trimToNBytes(b []byte, n int) []byte {
	if len(b) > n {
		return b[len(b)-n:]
	}

	if len(b) < n {
		return append(bytes.Repeat([]byte{0}, n-len(b)), b...)
	}

	return b
}

func (h Hex) Int64() (int64, error) {
	val, err := h.Uint64()
	if err != nil {
		return 0, err
	}
	return int64(val), err
}

func (h Hex) Uint() (uint, error) {
	val, err := h.Uint32()
	if err != nil {
		return 0, err
	}

	return uint(val), nil
}

func (h Hex) Time() (time.Time, error) {
	if len(h) == 0 {
		return time.Time{}, nil
	}

	hexStr := hex.EncodeToString(h)
	if hexStr == "" {
		return time.Time{}, nil
	}

	seconds, err := strconv.ParseUint(hexStr, 16, 64)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to parse hex timestamp")
	}

	return time.Unix(int64(seconds), 0).UTC(), nil
}

func (h Hex) Decimal() (decimal.Decimal, error) {
	if len(h) == 0 {
		return decimal.Zero, nil
	}

	hexStr := hex.EncodeToString(h)
	if hexStr == "" {
		return decimal.Zero, nil
	}

	val := new(big.Int)
	_, ok := val.SetString(hexStr, 16)
	if !ok {
		return decimal.Zero, errors.New("failed to parse hex to big.Int")
	}

	return decimal.NewFromBigInt(val, 0), nil
}

func (h Hex) BigInt() (*big.Int, error) {
	if len(h) == 0 {
		return big.NewInt(0), nil
	}

	hexStr := hex.EncodeToString(h)
	hexStr = strings.TrimPrefix(strings.ToLower(hexStr), "0x")
	if hexStr == "" {
		return big.NewInt(0), nil
	}

	val := new(big.Int)
	_, ok := val.SetString(hexStr, 16)
	if !ok {
		return nil, errors.Errorf("failed to parse hex to big.Int: %s", hexStr)
	}

	return val, nil
}

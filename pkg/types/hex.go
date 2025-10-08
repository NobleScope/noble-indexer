package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"github.com/pkg/errors"
	"strconv"
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

func (h Hex) String() string {
	if len(h) == 0 {
		return ""
	}
	hexStr := hex.EncodeToString(h)
	return "0x" + hexStr
}

func (h Hex) Uint64() (uint64, error) {
	if len(h) == 0 {
		return 0, nil
	}
	hexStr := hex.EncodeToString(h)
	val, err := strconv.ParseUint(hexStr, 16, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse hex to uint64")
	}
	return val, nil
}

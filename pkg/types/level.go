package types

import (
	"fmt"
	"strconv"
)

type Level uint64

func (l Level) String() string {
	return fmt.Sprintf("%d", l)
}

func (l Level) Hex() string {
	return "0x" + strconv.FormatUint(uint64(l), 16)
}

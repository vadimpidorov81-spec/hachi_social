package users

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type ID struct {
	value [16]byte
}

func ParseID(value string) (ID, error) {
	compact := strings.ReplaceAll(value, "-", "")
	if len(compact) != 32 {
		return ID{}, ErrInvalidID
	}

	decoded, err := hex.DecodeString(compact)
	if err != nil {
		return ID{}, ErrInvalidID
	}

	var id ID
	copy(id.value[:], decoded)
	if id.IsZero() {
		return ID{}, ErrInvalidID
	}
	return id, nil
}

func IDFromBytes(value [16]byte) ID {
	return ID{value: value}
}

func (id ID) Bytes() [16]byte {
	return id.value
}

func (id ID) IsZero() bool {
	return id == ID{}
}

func (id ID) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		id.value[0:4],
		id.value[4:6],
		id.value[6:8],
		id.value[8:10],
		id.value[10:16],
	)
}

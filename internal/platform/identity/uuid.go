package identity

import (
	"crypto/rand"
	"fmt"

	"github.com/hachisocial/hachisocial/internal/domain/users"
)

type UUIDv4Generator struct{}

func (UUIDv4Generator) New() (users.ID, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return users.ID{}, fmt.Errorf("generate uuid: %w", err)
	}

	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return users.IDFromBytes(value), nil
}

package users

import (
	"strings"
	"unicode/utf8"
)

const MaxUsernameLength = 20

var reservedUsernames = map[string]struct{}{
	"admin":         {},
	"administrator": {},
	"api":           {},
	"auth":          {},
	"bot":           {},
	"help":          {},
	"moderator":     {},
	"root":          {},
	"settings":      {},
	"support":       {},
	"system":        {},
	"users":         {},
}

type Username struct {
	value string
}

func NewUsername(value string) (Username, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || utf8.RuneCountInString(value) > MaxUsernameLength {
		return Username{}, ErrInvalidUsername
	}
	for _, character := range value {
		if character < 'a' || character > 'z' {
			return Username{}, ErrInvalidUsername
		}
	}
	if _, reserved := reservedUsernames[value]; reserved {
		return Username{}, ErrReservedUsername
	}
	return Username{value: value}, nil
}

func (u Username) String() string {
	return u.value
}

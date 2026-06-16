package users

import (
	"strings"
	"time"
	_ "time/tzdata"
	"unicode"
	"unicode/utf8"
)

const (
	MaxDisplayNameLength = 50
	MaxBioLength         = 500
)

type DisplayName struct {
	value string
}

func NewDisplayName(value string) (DisplayName, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > MaxDisplayNameLength {
		return DisplayName{}, ErrInvalidDisplayName
	}
	for _, character := range value {
		if character == '\n' || character == '\r' || unicode.IsControl(character) {
			return DisplayName{}, ErrInvalidDisplayName
		}
	}
	return DisplayName{value: value}, nil
}

func (d DisplayName) String() string {
	return d.value
}

type Bio struct {
	value string
}

func NewBio(value string) (Bio, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > MaxBioLength {
		return Bio{}, ErrInvalidBio
	}
	for _, character := range value {
		if unicode.IsControl(character) && character != '\n' && character != '\r' && character != '\t' {
			return Bio{}, ErrInvalidBio
		}
	}
	return Bio{value: value}, nil
}

func (b Bio) String() string {
	return b.value
}

type Timezone struct {
	value string
}

func NewTimezone(value string) (Timezone, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Timezone{}, ErrInvalidTimezone
	}
	if _, err := time.LoadLocation(value); err != nil {
		return Timezone{}, ErrInvalidTimezone
	}
	return Timezone{value: value}, nil
}

func (t Timezone) String() string {
	return t.value
}

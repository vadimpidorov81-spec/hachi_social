package users

import "errors"

var (
	ErrInvalidID          = errors.New("invalid user id")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrReservedUsername   = errors.New("reserved username")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrInvalidBio         = errors.New("invalid bio")
	ErrInvalidTimezone    = errors.New("invalid timezone")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidStatus      = errors.New("invalid status")
)

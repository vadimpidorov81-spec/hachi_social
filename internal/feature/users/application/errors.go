package application

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUsernameAlreadyTaken = errors.New("username already taken")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrCannotBlockSelf      = errors.New("administrator cannot block self")
	ErrEmptyUpdate          = errors.New("profile update is empty")
)

package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountDisabled    = errors.New("account is disabled. please contact admin")
	ErrUserNotFound       = errors.New("user not found")
)

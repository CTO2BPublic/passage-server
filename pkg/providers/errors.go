package providers

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrPermissionDenied = errors.New("permission denied")
)

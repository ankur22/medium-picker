package store

import "github.com/ankur22/medium-picker/internal/err"

const (
	ErrUserAlreadyExists = err.Const("user already exists")
	ErrUserNotFound      = err.Const("user not found")
)

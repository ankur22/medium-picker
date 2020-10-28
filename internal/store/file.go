package store

import "github.com/ankur22/medium-picker/internal/err"

const (
	ErrUserAlreadyExists         = err.Const("user already exists")
	ErrUserNotFound              = err.Const("user not found")
	ErrMediumSourceAlreadyExists = err.Const("medium source already exits")
)

type Source struct {
	URL string
	ID  string
}

package store

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ankur22/medium-picker/internal/err"
	"github.com/ankur22/medium-picker/internal/logging"
)

const (
	ErrCannotOpenUserFile        = err.Const("cannot open user file")
	ErrCannotReadUserFile        = err.Const("cannot read user file")
	ErrCannotUnmarshallUserFile  = err.Const("cannot unmarshall user file")
	ErrUserAlreadyExists         = err.Const("user already exists")
	ErrUserNotFound              = err.Const("user not found")
	ErrMediumSourceAlreadyExists = err.Const("medium source already exits")
)

type Source struct {
	URL string
	ID  string
}

// UserFile is the type that will store the user information in a file on disk
type UserFile struct {
	filename string
	ticker   time.Duration
	emails   map[string]string
	users    map[string]string
	lock     sync.Mutex
	dirty    bool
}

// NewUserFile will create a new instance of UserFile
// This is not thread safe
func NewUserFile(ctx context.Context, filename string, ticker time.Duration) (*UserFile, error) {
	u := UserFile{
		filename: filename,
		ticker:   ticker,
		emails:   make(map[string]string),
		users:    make(map[string]string),
	}

	if err := u.load(ctx); err != nil {
		return nil, err
	}

	return &u, nil
}

// CreateNewUser will create a new user that is needed to
// add medium sources. It returns the uuid of the new user.
func (u *UserFile) CreateNewUser(ctx context.Context, email string) (string, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if _, ok := u.emails[email]; ok {
		return "", ErrUserAlreadyExists
	}

	u.dirty = true
	u.emails[email] = uuid.New().String()
	u.users[u.emails[email]] = email

	return u.emails[email], nil
}

// GetUser will retrieve the user details. It will return
// the uuid of the user.
func (u *UserFile) GetUser(ctx context.Context, email string) (string, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if v, ok := u.emails[email]; ok {
		return v, nil
	}
	return "", ErrUserNotFound
}

// IsUser checks whether the given userID is valid
func (u *UserFile) IsUser(ctx context.Context, userID string) (bool, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if _, ok := u.users[userID]; ok {
		return true, nil
	}
	return false, nil
}

// Start will start the background job that will periodically save
// what's in memory
func (u *UserFile) Start(ctx context.Context) error {
	t := time.NewTicker(u.ticker)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}

		u.lock.Lock()
		defer u.lock.Unlock()

		if !u.dirty {
			continue
		}

		f, err := os.Create(u.filename)
		if err != nil {
			return ErrCannotOpenUserFile
		}
		defer func() {
			if err := f.Close(); err != nil {
				logging.Error(ctx, "cannot close user file", zap.Error(err))
			}
		}()

		data := userData{
			Emails: u.emails,
			Users:  u.users,
		}

		bb, err := json.Marshal(&data)
		if err != nil {
			logging.Error(ctx, "cannot marshal user data", zap.Error(err))
			continue
		}

		if _, err := f.Write(bb); err != nil {
			logging.Error(ctx, "cannot write user data", zap.Error(err))
			continue
		}

		u.dirty = false
	}
}

func (u *UserFile) load(ctx context.Context) error {
	if _, err := os.Stat(u.filename); os.IsNotExist(err) {
		return nil
	}

	f, err := os.Open(u.filename)
	if err != nil {
		return ErrCannotOpenUserFile
	}
	defer func() {
		if err := f.Close(); err != nil {
			logging.Error(ctx, "cannot close user file", zap.Error(err))
		}
	}()

	bb, err := ioutil.ReadAll(f)
	if err != nil {
		return ErrCannotReadUserFile
	}

	var data userData
	err = json.Unmarshal(bb, &data)
	if err != nil {
		return ErrCannotUnmarshallUserFile
	}

	u.emails = data.Emails
	u.users = data.Users

	return nil
}

type userData struct {
	Emails map[string]string `json:"emails"`
	Users  map[string]string `json:"users"`
}

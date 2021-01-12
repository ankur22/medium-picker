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
	ErrCannotOpenMediumFile       = err.Const("cannot open medium file")
	ErrCannotReadMediumFile       = err.Const("cannot read medium file")
	ErrCannotUnmarshallMediumFile = err.Const("cannot unmarshall medium file")
	ErrCannotFindMedium           = err.Const("cannot find medium")
)

// Source is the response type
type Source struct {
	URL string
	ID  string
}

type medium struct {
	URL          string    `json:"url"`
	ID           string    `json:"id"`
	Hash         string    `json:"hash"`
	Multiplier   float32   `json:"multiplier"`
	CreatedDate  time.Time `json:"created_date"`
	ModifiedDate time.Time `json:"modified_date"`
	Hit          int       `json:"hit"`
	UserID       string    `json:"user_id"`
}

// MediumFile is the type that will store the medium information in a file on disk
type MediumFile struct {
	filename    string
	ticker      time.Duration
	sources     map[string]map[string]medium
	lock        sync.Mutex
	dirty       bool
	elemsInPage int
}

// NewMediumFile will create a new instance of MediumFile
// This is not thread safe
func NewMediumFile(ctx context.Context, filename string, ticker time.Duration, elemsInPage int) (*MediumFile, error) {
	m := MediumFile{
		filename:    filename,
		ticker:      ticker,
		sources:     make(map[string]map[string]medium),
		elemsInPage: elemsInPage,
	}

	if err := m.load(ctx); err != nil {
		return nil, err
	}

	return &m, nil
}

// AddSource will add a new url medium source for a userID
func (m *MediumFile) AddSource(ctx context.Context, userID string, source string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	val, ok := m.sources[userID]
	if !ok {
		m.sources[userID] = make(map[string]medium)
		val = m.sources[userID]
	}

	if _, ok := val[source]; ok {
		return ErrMediumSourceAlreadyExists
	}

	val[source] = medium{
		URL:          source,
		ID:           uuid.New().String(),
		Hash:         "",
		Multiplier:   0,
		CreatedDate:  time.Now().UTC(),
		ModifiedDate: time.Now().UTC(),
		Hit:          0,
		UserID:       userID,
	}

	return nil
}

// GetSources returns the sources on the selected page
func (m *MediumFile) GetSources(ctx context.Context, userID string, page int) ([]Source, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	val, ok := m.sources[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	start := m.elemsInPage * page
	if start > len(val) {
		return nil, nil
	}

	end := start + m.elemsInPage
	if end > len(val) {
		end = len(val)
	}

	count := start
	var resp []Source
	for _, v := range val {
		count++

		if count < start {
			continue
		}

		if count > end {
			break
		}

		resp = append(resp, Source{
			URL: v.URL,
			ID:  v.ID,
		})
	}

	return resp, nil
}

// DeleteSource will delete a source given the userID and sourceID
func (m *MediumFile) DeleteSource(ctx context.Context, userID string, sourceID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	val, ok := m.sources[userID]
	if !ok {
		return ErrUserNotFound
	}

	var key string
	for k, v := range val {
		if v.ID == sourceID {
			key = k
			break
		}
	}

	if key == "" {
		return ErrCannotFindMedium
	}

	delete(val, key)

	return nil
}

// Start will start the background job that will periodically save
// what's in memory
func (m *MediumFile) Start(ctx context.Context) error {
	t := time.NewTicker(m.ticker)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}

		m.lock.Lock()
		defer m.lock.Unlock()

		if !m.dirty {
			continue
		}

		f, err := os.Create(m.filename)
		if err != nil {
			return ErrCannotOpenMediumFile.Wrap(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				logging.Error(ctx, "cannot close medium file", zap.Error(err))
			}
		}()

		bb, err := json.Marshal(&m.sources)
		if err != nil {
			logging.Error(ctx, "cannot marshal medium source data", zap.Error(err))
			continue
		}

		if _, err := f.Write(bb); err != nil {
			logging.Error(ctx, "cannot write medium source data", zap.Error(err))
			continue
		}

		m.dirty = false
	}
}

func (m *MediumFile) load(ctx context.Context) error {
	if _, err := os.Stat(m.filename); os.IsNotExist(err) {
		return nil
	}

	f, err := os.Open(m.filename)
	if err != nil {
		return ErrCannotOpenMediumFile.Wrap(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logging.Error(ctx, "cannot close medium source file", zap.Error(err))
		}
	}()

	bb, err := ioutil.ReadAll(f)
	if err != nil {
		return ErrCannotReadMediumFile.Wrap(err)
	}

	var data map[string]map[string]medium
	err = json.Unmarshal(bb, &data)
	if err != nil {
		return ErrCannotUnmarshallMediumFile.Wrap(err)
	}

	m.sources = data

	return nil
}

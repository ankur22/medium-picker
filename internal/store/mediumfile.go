package store

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/ankur22/medium-picker/internal/err"
	"github.com/ankur22/medium-picker/internal/logging"
	"go.uber.org/zap"
)

const (
	ErrCannotOpenMediumFile       = err.Const("cannot open medium file")
	ErrCannotReadMediumFile       = err.Const("cannot read medium file")
	ErrCannotUnmarshallMediumFile = err.Const("cannot unmarshall medium file")
)

type Source struct {
	URL string
	ID  string
}

// MediumFile is the type that will store the medium information in a file on disk
type MediumFile struct {
	filename string
	ticker   time.Duration
	sources  map[string][]Source
	lock     sync.Mutex
	dirty    bool
}

// NewMediumFile will create a new instance of MediumFile
// This is not thread safe
func NewMediumFile(ctx context.Context, filename string, ticker time.Duration) (*MediumFile, error) {
	m := MediumFile{
		filename: filename,
		ticker:   ticker,
		sources:  make(map[string][]Source),
	}

	if err := m.load(ctx); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *MediumFile) AddSource(ctx context.Context, userID string, source string) error {
	return errors.New("not implemented")
}

func (m *MediumFile) GetSources(ctx context.Context, userID string, page int) ([]store.Source, error) {
	return nil, errors.New("not implemented")
}

func (m *MediumFile) DeleteSource(ctx context.Context, userID string, sourceID string) error {
	return errors.New("not implemented")
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

	var data map[string][]Source
	err = json.Unmarshal(bb, &data)
	if err != nil {
		return ErrCannotUnmarshallMediumFile.Wrap(err)
	}

	m.sources = data

	return nil
}

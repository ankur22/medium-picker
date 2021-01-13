//go:generate mockgen -destination=mock_service.go -package=service github.com/ankur22/medium-picker/internal/service MediumSourceStorer

package service

import (
	"context"
	"sort"

	"github.com/ankur22/medium-picker/internal/err"
	"github.com/ankur22/medium-picker/internal/store"
)

const (
	ErrCountSmallerThanOne = err.Const("count is smaller than 1")
	ErrFailedGetAllSources = err.Const("failed to retrieve all records")
)

// MediumSourceStorer interface to retrieve medium sources
type MediumSourceStorer interface {
	GetAllSourceData(ctx context.Context, userID string, page int) ([]store.Medium, error)
	UpdateSource(ctx context.Context, userID string, source store.Medium) error
}

// Picker is where the main business logic of
// picking a source(s) to read for the user
type Picker struct {
	store MediumSourceStorer
}

// NewPicker will create a new instance of Picker
func NewPicker(store MediumSourceStorer) *Picker {
	return &Picker{
		store: store,
	}
}

// Pick will pick the source(s) for the user to read
func (p *Picker) Pick(ctx context.Context, userID string, count int) ([]store.Source, error) {
	if count < 1 {
		return nil, ErrCountSmallerThanOne
	}

	var page int
	sources := make([]store.Medium, 1)
	var all []store.Medium
	for len(sources) != 0 {
		ss, err := p.store.GetAllSourceData(ctx, userID, page)
		if err != nil {
			return nil, ErrFailedGetAllSources
		}
		all = append(all, ss...)
		page++
		sources = ss
	}

	sort.Slice(all, func(i, j int) bool {
		a := (float32(all[i].Hit) * all[i].Multiplier)
		b := (float32(all[j].Hit) * all[j].Multiplier)
		return a < b
	})

	if count > len(all) {
		count = len(all)
	}

	all = all[:count]

	sort.Slice(all, func(i, j int) bool {
		return all[i].ModifiedDate.Before(all[j].ModifiedDate)
	})

	rtnVal := make([]store.Source, 0, count)
	for i := 0; i < count; i++ {
		rtnVal = append(rtnVal, store.Source{
			URL: all[i].URL,
			ID:  all[i].ID,
		})
		all[i].Hit++
		p.store.UpdateSource(ctx, userID, all[i])
	}

	return rtnVal, nil
}

package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ankur22/medium-picker/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMediumFile_Success(t *testing.T) {
	m, err := store.NewMediumFile(context.Background(), "filename.json", time.Second, 10)
	assert.NoError(t, err)
	assert.NotNil(t, m)
}

func TestMediumFile_AddSource(t *testing.T) {
	type args struct {
		userID string
		source string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Success",
			args: args{
				userID: "some-user-id",
				source: "google.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m, err := store.NewMediumFile(ctx, "filename.json", time.Second, 10)
			require.NoError(t, err)
			require.NotNil(t, m)

			err = m.AddSource(ctx, tt.args.userID, tt.args.source)
			assert.NoError(t, err)
		})
	}
}

func TestMediumFile_DeleteSource(t *testing.T) {
	type fields struct {
		source string
	}
	type args struct {
		userID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "successfully deleted the new medium",
			fields: fields{
				source: "google.com",
			},
			args: args{
				userID: "some-user-id",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m, err := store.NewMediumFile(ctx, "filename.json", time.Second, 1)
			require.NoError(t, err)
			require.NotNil(t, m)

			err = m.AddSource(ctx, tt.args.userID, tt.fields.source)
			require.NoError(t, err)

			sources, err := m.GetSources(ctx, tt.args.userID, 0)
			require.NoError(t, err)

			err = m.DeleteSource(ctx, tt.args.userID, sources[0].ID)
			assert.NoError(t, err)
		})
	}
}

func TestMediumFile_GetSources(t *testing.T) {
	type fields struct {
		source string
		count  int
	}
	type args struct {
		userID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Retrieved all sources",
			fields: fields{
				source: "google.com",
				count:  20,
			},
			args: args{
				userID: "some-user-id",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m, err := store.NewMediumFile(ctx, "filename.json", time.Second, 5)
			require.NoError(t, err)
			require.NotNil(t, m)

			for i := 0; i < tt.fields.count; i++ {
				err = m.AddSource(ctx, tt.args.userID, fmt.Sprintf("%d%s", i, tt.fields.source))
				require.NoError(t, err)
			}

			var page int
			sources := make([]store.Source, 0)
			var got []store.Source
			for sources != nil {
				sources, err = m.GetSources(ctx, tt.args.userID, page)
				assert.NoError(t, err)
				got = append(got, sources...)
				page++
			}

			assert.Equal(t, tt.fields.count, len(got))
		})
	}
}

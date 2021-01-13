package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ankur22/medium-picker/internal/service"
	"github.com/ankur22/medium-picker/internal/store"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewPicker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := service.NewMockMediumSourceStorer(ctrl)

	p := service.NewPicker(s)
	assert.NotNil(t, p)
}

func TestPicker_Pick(t *testing.T) {
	type fields struct {
		sources []store.Medium
	}
	type args struct {
		userID string
		count  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []store.Source
		wantHit []int
	}{
		{
			name: "success with 1",
			fields: fields{
				sources: []store.Medium{
					store.Medium{
						URL:          "a.com",
						ID:           "1",
						Hit:          10,
						Multiplier:   1,
						ModifiedDate: time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC),
					},
					store.Medium{
						URL:          "b.com",
						ID:           "2",
						Hit:          5,
						Multiplier:   1,
						ModifiedDate: time.Date(0, 0, 0, 1, 0, 0, 0, time.UTC),
					},
					store.Medium{
						URL:          "c.com",
						ID:           "3",
						Hit:          15,
						Multiplier:   0.01,
						ModifiedDate: time.Date(0, 0, 0, 5, 0, 0, 0, time.UTC),
					},
					store.Medium{
						URL:          "d.com",
						ID:           "4",
						Hit:          1,
						Multiplier:   1,
						ModifiedDate: time.Date(0, 0, 0, 16, 0, 0, 0, time.UTC),
					},
				},
			},
			args: args{
				userID: "some-id",
				count:  1,
			},
			want: []store.Source{
				store.Source{
					URL: "c.com",
					ID:  "3",
				},
			},
			wantHit: []int{16},
		},
		{
			name: "success with 2",
			fields: fields{
				sources: []store.Medium{
					store.Medium{
						URL:          "a.com",
						ID:           "1",
						Hit:          10,
						Multiplier:   1,
						ModifiedDate: time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC),
					},
					store.Medium{
						URL:          "b.com",
						ID:           "2",
						Hit:          5,
						Multiplier:   1,
						ModifiedDate: time.Date(0, 0, 0, 1, 0, 0, 0, time.UTC),
					},
					store.Medium{
						URL:          "c.com",
						ID:           "3",
						Hit:          15,
						Multiplier:   1,
						ModifiedDate: time.Date(0, 0, 0, 5, 0, 0, 0, time.UTC),
					},
					store.Medium{
						URL:          "d.com",
						ID:           "4",
						Hit:          0,
						Multiplier:   1,
						ModifiedDate: time.Date(0, 0, 0, 16, 0, 0, 0, time.UTC),
					},
				},
			},
			args: args{
				userID: "some-id",
				count:  2,
			},
			want: []store.Source{
				store.Source{
					URL: "b.com",
					ID:  "2",
				},
				store.Source{
					URL: "d.com",
					ID:  "4",
				},
			},
			wantHit: []int{6, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var count int
			s := service.NewMockMediumSourceStorer(ctrl)
			s.EXPECT().GetAllSourceData(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, userID string, page int) ([]store.Medium, error) {
				count++
				if count == 1 {
					return tt.fields.sources, nil
				}
				return nil, nil
			}).Times(2)

			var index int
			s.EXPECT().UpdateSource(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, userID string, source store.Medium) error {
				assert.Equal(t, tt.want[index].ID, source.ID)
				assert.Equal(t, tt.want[index].URL, source.URL)
				assert.Equal(t, tt.wantHit[index], source.Hit)
				index++
				return nil
			}).Times(len(tt.wantHit))

			p := service.NewPicker(s)

			ss, err := p.Pick(ctx, tt.args.userID, tt.args.count)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.want, ss)
		})
	}
}

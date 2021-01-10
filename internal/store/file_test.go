package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankur22/medium-picker/internal/store"
)

func TestNewUserFile_Success(t *testing.T) {
	u, err := store.NewUserFile(context.Background(), "some-file.txt", time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestUserFile_CreateNewUser_Success(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "User created",
			args: args{
				email: "test@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := store.NewUserFile(context.Background(), "some-file.txt", time.Second)
			require.NoError(t, err)

			uid, err := u.CreateNewUser(context.Background(), tt.args.email)
			assert.NoError(t, err)
			assert.NotEmpty(t, uid)
		})
	}
}

func TestUserFile_GetUser_Success(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "User created",
			args: args{
				email: "test@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := store.NewUserFile(context.Background(), "some-file.txt", time.Second)
			require.NoError(t, err)

			uid, err := u.CreateNewUser(context.Background(), tt.args.email)
			require.NoError(t, err)
			require.NotEmpty(t, uid)

			uid, err = u.GetUser(context.Background(), tt.args.email)
			assert.NoError(t, err)
			assert.NotEmpty(t, uid)
		})
	}
}

func TestUserFile_GetUser_Failure(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "User created",
			args: args{
				email: "test@example.com",
			},
			wantErr: store.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := store.NewUserFile(context.Background(), "some-file.txt", time.Second)
			require.NoError(t, err)

			uid, err := u.GetUser(context.Background(), tt.args.email)
			assert.Empty(t, uid)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func TestUserFile_IsUser_Success(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "User exists",
			args: args{
				email: "test@example.com",
			},
			want: true,
		},
		{
			name: "User doesn't exist",
			args: args{
				email: "test@example.com",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := store.NewUserFile(context.Background(), "some-file.txt", time.Second)
			require.NoError(t, err)

			var uid string
			if tt.want {
				uid, err = u.CreateNewUser(context.Background(), tt.args.email)
				require.NoError(t, err)
				require.NotEmpty(t, uid)
			}

			got, err := u.IsUser(context.Background(), uid)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

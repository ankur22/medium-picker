package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/ankur22/medium-picker/internal/logging"
	"github.com/ankur22/medium-picker/internal/rest"
	"github.com/ankur22/medium-picker/internal/store"
	pkgRest "github.com/ankur22/medium-picker/pkg/rest"
)

func Test_Handler_Signup_Success(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name   string
		body   interface{}
		email  string
		userID string
	}{
		{
			name:   "Signup successful",
			body:   pkgRest.SignupRequest{Email: "test@email.com"},
			email:  "test@email.com",
			userID: "a09sd09sa8d0a8sd",
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStore(ctrl)
		s.EXPECT().CreateNewUser(gomock.Any(), tt.email).Return(tt.userID, nil)

		h := rest.NewHandler(s, nil)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqB))

		h.Signup(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Result().StatusCode)

		b, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)

		respB := pkgRest.SignupResponse{}

		err = json.Unmarshal(b, &respB)
		assert.NoError(t, err)

		assert.Equal(t, tt.userID, respB.UserID)
	}
}

func Test_Handler_Signup_Failure(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name          string
		body          interface{}
		expectedError int
		storeError    error
	}{
		{
			name:          "No body",
			body:          nil,
			expectedError: http.StatusBadRequest,
			storeError:    nil,
		},
		{
			name:          "Invalid email",
			body:          pkgRest.SignupRequest{Email: "not an email"},
			expectedError: http.StatusBadRequest,
			storeError:    nil,
		},
		{
			name:          "Store failed",
			body:          pkgRest.SignupRequest{Email: "test@email.com"},
			expectedError: http.StatusInternalServerError,
			storeError:    errors.New("some error"),
		},
		{
			name:          "User already exists",
			body:          pkgRest.SignupRequest{Email: "test@email.com"},
			expectedError: http.StatusConflict,
			storeError:    store.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStore(ctrl)
		if tt.storeError != nil {
			s.EXPECT().CreateNewUser(gomock.Any(), gomock.Any()).Return("", tt.storeError)
		}

		h := rest.NewHandler(s, nil)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqB))

		h.Signup(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

func Test_Handler_SignIn_Success(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name   string
		body   interface{}
		email  string
		userID string
	}{
		{
			name:   "Sign in successful",
			body:   pkgRest.SignInRequest{Email: "test@email.com"},
			email:  "test@email.com",
			userID: "a09sd09sa8d0a8sd",
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStore(ctrl)
		s.EXPECT().GetUser(gomock.Any(), tt.email).Return(tt.userID, nil)

		h := rest.NewHandler(s, nil)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/", bytes.NewBuffer(reqB))

		h.SignIn(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		b, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)

		respB := pkgRest.SignInResponse{}

		err = json.Unmarshal(b, &respB)
		assert.NoError(t, err)

		assert.Equal(t, tt.userID, respB.UserID)
	}
}

func Test_Handler_SignIn_Failure(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name          string
		body          interface{}
		expectedError int
		storeError    error
	}{
		{
			name:          "No body",
			body:          nil,
			expectedError: http.StatusBadRequest,
			storeError:    nil,
		},
		{
			name:          "Invalid email",
			body:          pkgRest.SignInRequest{Email: "not an email"},
			expectedError: http.StatusBadRequest,
			storeError:    nil,
		},
		{
			name:          "Store failed",
			body:          pkgRest.SignInRequest{Email: "test@email.com"},
			expectedError: http.StatusInternalServerError,
			storeError:    errors.New("some error"),
		},
		{
			name:          "User not found",
			body:          pkgRest.SignInRequest{Email: "test@email.com"},
			expectedError: http.StatusNotFound,
			storeError:    store.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStore(ctrl)
		if tt.storeError != nil {
			s.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return("", tt.storeError)
		}

		h := rest.NewHandler(s, nil)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/", bytes.NewBuffer(reqB))

		h.SignIn(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

func Test_Handler_AddMediumSource_Success(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name   string
		body   interface{}
		userID string
		source string
	}{
		{
			name:   "Add source successful",
			body:   pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID: "ds098fa0s98fd0sa",
			source: "google.com/news",
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStore(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(true, nil)

		m := rest.NewMockMediumSourceStore(ctrl)
		m.EXPECT().AddSource(gomock.Any(), tt.userID, tt.source).Return(nil)

		h := rest.NewHandler(s, m)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqB))
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID})

		h.AddMediumSource(resp, req)

		assert.Equal(t, http.StatusNoContent, resp.Result().StatusCode)
	}
}

func Test_Handler_AddMediumSource_Failure(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name           string
		body           interface{}
		userID         string
		expectedError  int
		userFound      bool
		userStoreError error
		sourceError    error
	}{
		{
			name:           "User not found",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			expectedError:  http.StatusNotFound,
			userFound:      false,
			userStoreError: nil,
		},
		{
			name:           "User store errors",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			expectedError:  http.StatusInternalServerError,
			userFound:      false,
			userStoreError: errors.New("some error"),
		},
		{
			name:           "Source already exists",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			expectedError:  http.StatusConflict,
			userFound:      true,
			userStoreError: nil,
			sourceError:    store.ErrMediumSourceAlreadyExists,
		},
		{
			name:           "Source store error",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			expectedError:  http.StatusInternalServerError,
			userFound:      true,
			userStoreError: nil,
			sourceError:    errors.New("some error"),
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStore(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(tt.userFound, tt.userStoreError)

		m := rest.NewMockMediumSourceStore(ctrl)
		if tt.userFound {
			m.EXPECT().AddSource(gomock.Any(), tt.userID, gomock.Any()).Return(tt.sourceError)
		}

		h := rest.NewHandler(s, m)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqB))
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID})

		h.AddMediumSource(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

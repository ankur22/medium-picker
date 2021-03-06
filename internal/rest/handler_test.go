package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/ankur22/medium-picker/internal/logging"
	"github.com/ankur22/medium-picker/internal/rest"
	"github.com/ankur22/medium-picker/internal/store"
	pkgRest "github.com/ankur22/medium-picker/pkg/rest"
)

func TestHandler_Signup_Success(t *testing.T) {
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

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().CreateNewUser(gomock.Any(), tt.email).Return(tt.userID, nil)

		h := rest.NewHandler(s, nil, nil)

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

func TestHandler_Signup_Failure(t *testing.T) {
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

		s := rest.NewMockUserStorer(ctrl)
		if tt.storeError != nil {
			s.EXPECT().CreateNewUser(gomock.Any(), gomock.Any()).Return("", tt.storeError)
		}

		h := rest.NewHandler(s, nil, nil)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqB))

		h.Signup(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

func TestHandler_SignIn_Success(t *testing.T) {
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

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().GetUser(gomock.Any(), tt.email).Return(tt.userID, nil)

		h := rest.NewHandler(s, nil, nil)

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

func TestHandler_SignIn_Failure(t *testing.T) {
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

		s := rest.NewMockUserStorer(ctrl)
		if tt.storeError != nil {
			s.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return("", tt.storeError)
		}

		h := rest.NewHandler(s, nil, nil)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/", bytes.NewBuffer(reqB))

		h.SignIn(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

func TestHandler_AddMediumSource_Success(t *testing.T) {
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

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(true, nil)

		m := rest.NewMockMediumSourceStorer(ctrl)
		m.EXPECT().AddSource(gomock.Any(), tt.userID, tt.source).Return(nil)

		h := rest.NewHandler(s, m, nil)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqB))
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID})

		h.AddMediumSource(resp, req)

		assert.Equal(t, http.StatusNoContent, resp.Result().StatusCode)
	}
}

func TestHandler_AddMediumSource_Failure(t *testing.T) {
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

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(tt.userFound, tt.userStoreError)

		m := rest.NewMockMediumSourceStorer(ctrl)
		if tt.userFound {
			m.EXPECT().AddSource(gomock.Any(), tt.userID, gomock.Any()).Return(tt.sourceError)
		}

		h := rest.NewHandler(s, m, nil)

		reqB, err := json.Marshal(tt.body)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqB))
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID})

		h.AddMediumSource(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

func TestHandler_GetMediumSource_Success(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name        string
		page        int
		userID      string
		storeResult []store.Source
	}{
		{
			name:   "Get sources",
			page:   0,
			userID: "ds098fa0s98fd0sa",
			storeResult: []store.Source{
				{ID: "1", URL: "google.com"}, {ID: "2", URL: "yahoo.com"},
			},
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(true, nil)

		m := rest.NewMockMediumSourceStorer(ctrl)
		m.EXPECT().GetSources(gomock.Any(), tt.userID, tt.page).Return(tt.storeResult, nil)

		h := rest.NewHandler(s, m, nil)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID, "page": strconv.Itoa(tt.page)})

		h.GetMediumSource(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
		defer resp.Result().Body.Close()

		bs, err := ioutil.ReadAll(resp.Result().Body)
		assert.NoError(t, err)

		var rBody []store.Source
		err = json.Unmarshal(bs, &rBody)
		assert.NoError(t, err)
		assert.ElementsMatch(t, tt.storeResult, rBody)
	}
}

func TestHandler_GetMediumSource_Failure(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name           string
		body           interface{}
		userID         string
		page           int
		expectedError  int
		userFound      bool
		userStoreError error
		sourceError    error
	}{
		{
			name:           "User not found",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			page:           0,
			expectedError:  http.StatusNotFound,
			userFound:      false,
			userStoreError: nil,
		},
		{
			name:           "User store errors",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			page:           0,
			expectedError:  http.StatusInternalServerError,
			userFound:      false,
			userStoreError: errors.New("some error"),
		},
		{
			name:           "page less than 0",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			page:           -1,
			expectedError:  http.StatusBadRequest,
			userFound:      true,
			userStoreError: nil,
			sourceError:    nil,
		},
		{
			name:           "Source store error",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			page:           0,
			expectedError:  http.StatusInternalServerError,
			userFound:      true,
			userStoreError: nil,
			sourceError:    errors.New("some error"),
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(tt.userFound, tt.userStoreError)

		m := rest.NewMockMediumSourceStorer(ctrl)
		if tt.userFound && tt.sourceError != nil {
			m.EXPECT().GetSources(gomock.Any(), tt.userID, tt.page).Return(nil, tt.sourceError)
		}

		h := rest.NewHandler(s, m, nil)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID, "page": strconv.Itoa(tt.page)})

		h.GetMediumSource(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

func TestHandler_DeleteMediumSource_Success(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name     string
		sourceID string
		userID   string
	}{
		{
			name:     "Get sources",
			sourceID: "kjn4t43wknt",
			userID:   "ds098fa0s98fd0sa",
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(true, nil)

		m := rest.NewMockMediumSourceStorer(ctrl)
		m.EXPECT().DeleteSource(gomock.Any(), tt.userID, tt.sourceID).Return(nil)

		h := rest.NewHandler(s, m, nil)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID, "sourceID": tt.sourceID})

		h.DeleteMediumSource(resp, req)

		assert.Equal(t, http.StatusNoContent, resp.Result().StatusCode)
	}
}

func TestHandler_DeleteMediumSource_Failure(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name           string
		userID         string
		sourceID       string
		expectedError  int
		userFound      bool
		userStoreError error
		sourceError    error
	}{
		{
			name:           "User not found",
			userID:         "ds098fa0s98fd0sa",
			sourceID:       "32j4234j2oi3",
			expectedError:  http.StatusNotFound,
			userFound:      false,
			userStoreError: nil,
		},
		{
			name:           "User store errors",
			userID:         "ds098fa0s98fd0sa",
			sourceID:       "32j4234j2oi3",
			expectedError:  http.StatusInternalServerError,
			userFound:      false,
			userStoreError: errors.New("some error"),
		},
		{
			name:           "Source store error",
			userID:         "ds098fa0s98fd0sa",
			sourceID:       "32j4234j2oi3",
			expectedError:  http.StatusInternalServerError,
			userFound:      true,
			userStoreError: nil,
			sourceError:    errors.New("some error"),
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(tt.userFound, tt.userStoreError)

		m := rest.NewMockMediumSourceStorer(ctrl)
		if tt.userFound {
			m.EXPECT().DeleteSource(gomock.Any(), tt.userID, tt.sourceID).Return(tt.sourceError)
		}

		h := rest.NewHandler(s, m, nil)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID, "sourceID": tt.sourceID})

		h.DeleteMediumSource(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

func TestHandler_PickSources_Success(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name        string
		count       int
		userID      string
		storeResult []store.Source
	}{
		{
			name:   "Pick sources",
			count:  0,
			userID: "ds098fa0s98fd0sa",
			storeResult: []store.Source{
				{ID: "1", URL: "google.com"}, {ID: "2", URL: "yahoo.com"},
			},
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(true, nil)

		p := rest.NewMockMediumSourcePicker(ctrl)
		p.EXPECT().Pick(gomock.Any(), tt.userID, tt.count).Return(tt.storeResult, nil)

		h := rest.NewHandler(s, nil, p)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID, "count": strconv.Itoa(tt.count)})

		h.PickSources(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
		defer resp.Result().Body.Close()

		bs, err := ioutil.ReadAll(resp.Result().Body)
		assert.NoError(t, err)

		var rBody []store.Source
		err = json.Unmarshal(bs, &rBody)
		assert.NoError(t, err)
		assert.ElementsMatch(t, tt.storeResult, rBody)
	}
}

func TestHandler_PickSources_Failure(t *testing.T) {
	_, _ = logging.TestContext(context.Background())

	tests := []struct {
		name           string
		body           interface{}
		userID         string
		count          int
		expectedError  int
		userFound      bool
		userStoreError error
		sourceError    error
	}{
		{
			name:           "User not found",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			count:          0,
			expectedError:  http.StatusNotFound,
			userFound:      false,
			userStoreError: nil,
		},
		{
			name:           "User store errors",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			count:          0,
			expectedError:  http.StatusInternalServerError,
			userFound:      false,
			userStoreError: errors.New("some error"),
		},
		{
			name:           "count less than 0",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			count:          -1,
			expectedError:  http.StatusBadRequest,
			userFound:      true,
			userStoreError: nil,
			sourceError:    nil,
		},
		{
			name:           "Source store error",
			body:           pkgRest.NewMediumSourceRequest{Source: "google.com/news"},
			userID:         "ds098fa0s98fd0sa",
			count:          0,
			expectedError:  http.StatusInternalServerError,
			userFound:      true,
			userStoreError: nil,
			sourceError:    errors.New("some error"),
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := rest.NewMockUserStorer(ctrl)
		s.EXPECT().IsUser(gomock.Any(), tt.userID).Return(tt.userFound, tt.userStoreError)

		p := rest.NewMockMediumSourcePicker(ctrl)
		if tt.userFound && tt.sourceError != nil {
			p.EXPECT().Pick(gomock.Any(), tt.userID, tt.count).Return(nil, tt.sourceError)
		}

		h := rest.NewHandler(s, nil, p)

		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"userID": tt.userID, "count": strconv.Itoa(tt.count)})

		h.PickSources(resp, req)

		assert.Equal(t, tt.expectedError, resp.Result().StatusCode)
	}
}

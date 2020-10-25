package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ankur22/medium-picker/internal/logging"
	"github.com/ankur22/medium-picker/internal/rest"
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

		s := rest.NewMockStore(ctrl)
		s.EXPECT().CreateNewUser(gomock.Any(), tt.email).Return(tt.userID, nil)

		h := rest.NewHandler(s)

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

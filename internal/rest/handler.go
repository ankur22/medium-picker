//go:generate mockgen -destination=mock_store.go -package=rest github.com/ankur22/medium-picker/internal/rest Store

package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/ankur22/medium-picker/internal/logging"
	"github.com/ankur22/medium-picker/internal/store"
	pkgRest "github.com/ankur22/medium-picker/pkg/rest"
)

type Store interface {
	CreateNewUser(ctx context.Context, email string) (string, error)
	GetUser(ctx context.Context, email string) (string, error)
}

type Handler struct {
	s Store
}

func NewHandler(s Store) *Handler {
	return &Handler{s: s}
}

func (h *Handler) Add(r *mux.Router) {
	r.HandleFunc("/user/{userID}", h.Signup)
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()

	rb := pkgRest.SignupRequest{}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Error(ctx, "Can't read body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(b, &rb)
	if err != nil {
		logging.Error(ctx, "Can't unmarshall body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !emailRegex.MatchString(rb.Email) {
		logging.Error(ctx, "Email failed validation")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := h.s.CreateNewUser(ctx, rb.Email)
	if errors.Is(err, store.ErrUserAlreadyExists) {
		logging.Info(ctx, "User already exists")
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		logging.Error(ctx, "Failed to store new user details", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: Check for temporary behaviour errors and retry in the store itself
	respB := pkgRest.SignupResponse{UserID: id}
	b, err = json.Marshal(respB)
	if err != nil {
		logging.Error(ctx, "failed to marshall signup response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(b)
	if err != nil {
		logging.Error(ctx, "failed to write signup response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()

	rb := pkgRest.SignInRequest{}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Error(ctx, "Can't read body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(b, &rb)
	if err != nil {
		logging.Error(ctx, "Can't unmarshall body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !emailRegex.MatchString(rb.Email) {
		logging.Error(ctx, "Email failed validation")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := h.s.GetUser(ctx, rb.Email)
	if errors.Is(err, store.ErrUserNotFound) {
		logging.Info(ctx, "User not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		logging.Error(ctx, "Failed to retrieve user from the store", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: Check for temporary behaviour errors and retry in the store itself
	respB := pkgRest.SignInResponse{UserID: id}
	b, err = json.Marshal(respB)
	if err != nil {
		logging.Error(ctx, "failed to marshall sign in response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		logging.Error(ctx, "failed to write sign in response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

var emailRegex = regexp.MustCompile(`(?:[a-z0-9!#$%&'*+/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)

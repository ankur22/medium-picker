//go:generate mockgen -destination=mock_store.go -package=rest github.com/ankur22/medium-picker/internal/rest UserStore,MediumSourceStore

package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/ankur22/medium-picker/internal/logging"
	"github.com/ankur22/medium-picker/internal/store"
	pkgRest "github.com/ankur22/medium-picker/pkg/rest"
)

// UserStore interface that will be used to retrieve user details
type UserStore interface {
	CreateNewUser(ctx context.Context, email string) (string, error)
	GetUser(ctx context.Context, email string) (string, error)
	IsUser(ctx context.Context, userID string) (bool, error)
}

// MediumSourceStore interface to retrieve medium sources
type MediumSourceStore interface {
	AddSource(ctx context.Context, userID string, source string) error
	GetSources(ctx context.Context, userID string, page int) ([]store.Source, error)
	DeleteSource(ctx context.Context, userID string, sourceID string) error
}

// Handler type for the REST service's endpoints
type Handler struct {
	s UserStore
	m MediumSourceStore
}

// NewHandler creates a new handler
// The stores cannot be nil
func NewHandler(s UserStore, m MediumSourceStore) *Handler {
	return &Handler{s: s, m: m}
}

// Add will wire up the endpoints to the handler methods
func (h *Handler) Add(r *mux.Router) {
	r.HandleFunc("/v1/user", h.Signup).Methods("POST")
	r.HandleFunc("/v1/user/login", h.SignIn).Methods("PUT")
	r.HandleFunc("/v1/user/{userID}/medium", h.AddMediumSource).Methods("POST")
	r.HandleFunc("/v1/user/{userID}/medium", h.GetMediumSource).Methods("GET").Queries("p", "{page:[0-9]+}")
	r.HandleFunc("/v1/user/{userID}/medium/{sourceID}", h.DeleteMediumSource).Methods("DELETE")
}

// Signup is the handler that will create a new user
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
		w.WriteHeader(http.StatusInternalServerError)
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

	logging.Info(ctx, "User signed up", zap.String("userId", id))
}

// SignIn will sign a existing user in
// Currently just returns the userID for the user, no real authentication
// TODO: Use vault auth
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
		w.WriteHeader(http.StatusInternalServerError)
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

	logging.Info(ctx, "User signed in", zap.String("userId", id))
}

// AddMediumSource will add a new medium source for the userID
func (h *Handler) AddMediumSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()

	pathParam := mux.Vars(r)
	userID := pathParam["userID"]

	ctx = logging.With(ctx, zap.String("userId", userID))

	if err := h.isUser(ctx, userID, w); err != nil {
		return
	}

	rb := pkgRest.NewMediumSourceRequest{}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Error(ctx, "Can't read body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(b, &rb)
	if err != nil {
		logging.Error(ctx, "Can't unmarshall body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.m.AddSource(ctx, userID, rb.Source)
	if errors.Is(err, store.ErrMediumSourceAlreadyExists) {
		logging.Info(ctx, "Medium source already exists")
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		logging.Error(ctx, "Error from store")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logging.Info(ctx, "Added a new medium source")
	w.WriteHeader(http.StatusNoContent)
}

// GetMediumSource retrieves all the medium sources for the specified userID
// The response is paginated and if nextPage in the response is non-nil then keep performing the request with query p=value of nextPage
func (h *Handler) GetMediumSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()

	params := mux.Vars(r)
	userID := params["userID"]
	page := params["page"]

	logging.Info(ctx, "", zap.String("page", page))

	ctx = logging.With(ctx, zap.String("userId", userID))

	if err := h.isUser(ctx, userID, w); err != nil {
		return
	}

	p, err := strconv.ParseInt(page, 10, 32)
	if err != nil {
		logging.Error(ctx, "Page query cannot be parsed to int")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if p < 0 {
		logging.Info(ctx, "Page query is less than 0", zap.Int("page", int(p)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srcs, err := h.m.GetSources(ctx, userID, int(p))
	if err != nil {
		logging.Error(ctx, "Error from store")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make([]pkgRest.Source, len(srcs))
	for i, s := range srcs {
		resp[i] = pkgRest.Source{
			ID:  s.ID,
			URL: s.URL,
		}
	}

	respB, err := json.Marshal(resp)
	if err != nil {
		logging.Error(ctx, "failed to marshall get sources response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respB)
	if err != nil {
		logging.Error(ctx, "failed to write get sources response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// DeleteMediumSource deletes the source for the specified userID with sourceID
func (h *Handler) DeleteMediumSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()

	params := mux.Vars(r)
	userID := params["userID"]
	sourceID := params["sourceID"]

	ctx = logging.With(ctx, zap.String("userId", userID), zap.String("sourceID", sourceID))

	if err := h.isUser(ctx, userID, w); err != nil {
		return
	}

	if err := h.m.DeleteSource(ctx, userID, sourceID); err != nil {
		logging.Error(ctx, "Error from store")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var emailRegex = regexp.MustCompile(`(?:[a-z0-9!#$%&'*+/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)

func (h *Handler) isUser(ctx context.Context, userID string, w http.ResponseWriter) error {
	if ok, err := h.s.IsUser(ctx, userID); err != nil {
		logging.Error(ctx, "Error from store")
		w.WriteHeader(http.StatusInternalServerError)
		return err
	} else if !ok {
		logging.Info(ctx, "UserID not found")
		w.WriteHeader(http.StatusNotFound)
		return errors.New("user with userID not found")
	}
	return nil
}

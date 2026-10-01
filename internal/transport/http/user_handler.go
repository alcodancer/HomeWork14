package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"example/internal/domain"
	"example/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u, err := h.svc.CreateUser(r.Context(), req.Name, req.Email)
	if err != nil {
		writeErr(w, err)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	u, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u, err := h.svc.UpdateUser(r.Context(), req.ID, req.Name, req.Email)
	if err != nil {
		writeErr(w, err)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) History(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	entries, err := h.svc.GetHistory(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	json.NewEncoder(w).Encode(entries)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidName), errors.Is(err, domain.ErrInvalidEmail):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrUserExists):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
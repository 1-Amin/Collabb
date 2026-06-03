package auth

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/auth/register", h.register).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", h.login).Methods(http.MethodPost)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	out, err := h.svc.Register(r.Context(), in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(out)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	out, err := h.svc.Login(r.Context(), in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

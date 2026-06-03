package board

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/collabb/backend/internal/auth"
	"github.com/collabb/backend/internal/ws"
	"github.com/gorilla/mux"
)

type Handler struct {
	svc *Service
	hub *ws.Hub
}

func NewHandler(svc *Service, hub *ws.Hub) *Handler {
	return &Handler{svc: svc, hub: hub}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/boards", h.createBoard).Methods(http.MethodPost)
	r.HandleFunc("/boards", h.listBoards).Methods(http.MethodGet)
	r.HandleFunc("/boards/{boardID}", h.getBoard).Methods(http.MethodGet)

	r.HandleFunc("/boards/{boardID}/columns", h.createColumn).Methods(http.MethodPost)
	r.HandleFunc("/columns/{colID}", h.updateColumn).Methods(http.MethodPut)
	r.HandleFunc("/columns/{colID}", h.deleteColumn).Methods(http.MethodDelete)

	r.HandleFunc("/columns/{colID}/cards", h.createCard).Methods(http.MethodPost)
	r.HandleFunc("/cards/{cardID}", h.updateCard).Methods(http.MethodPut)
	r.HandleFunc("/cards/{cardID}/move", h.moveCard).Methods(http.MethodPatch)
	r.HandleFunc("/cards/{cardID}", h.deleteCard).Methods(http.MethodDelete)
}

func userID(r *http.Request) string {
	v, _ := r.Context().Value(auth.UserIDKey).(string)
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) broadcast(boardID, eventType string, payload any) {
	data, _ := json.Marshal(payload)
	h.hub.Broadcast(ws.Message{BoardID: boardID, Type: eventType, Payload: data})
}

// ── Boards ───────────────────────────────────────────────────────────────────

func (h *Handler) createBoard(w http.ResponseWriter, r *http.Request) {
	var in struct{ Title string `json:"title"` }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	b, err := h.svc.CreateBoard(r.Context(), userID(r), in.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) listBoards(w http.ResponseWriter, r *http.Request) {
	boards, err := h.svc.ListBoards(r.Context(), userID(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, boards)
}

func (h *Handler) getBoard(w http.ResponseWriter, r *http.Request) {
	b, err := h.svc.GetBoard(r.Context(), mux.Vars(r)["boardID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// ── Columns ──────────────────────────────────────────────────────────────────

func (h *Handler) createColumn(w http.ResponseWriter, r *http.Request) {
	boardID := mux.Vars(r)["boardID"]
	var in struct{ Title string `json:"title"` }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	col, err := h.svc.CreateColumn(r.Context(), boardID, in.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broadcast(boardID, "column.created", col)
	writeJSON(w, http.StatusCreated, col)
}

func (h *Handler) updateColumn(w http.ResponseWriter, r *http.Request) {
	var in struct{ Title string `json:"title"` }
	json.NewDecoder(r.Body).Decode(&in)
	col, err := h.svc.UpdateColumn(r.Context(), mux.Vars(r)["colID"], in.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.broadcast(col.BoardID, "column.updated", col)
	writeJSON(w, http.StatusOK, col)
}

func (h *Handler) deleteColumn(w http.ResponseWriter, r *http.Request) {
	colID := mux.Vars(r)["colID"]
	if err := h.svc.DeleteColumn(r.Context(), colID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Cards ─────────────────────────────────────────────────────────────────────

func (h *Handler) createCard(w http.ResponseWriter, r *http.Request) {
	colID := mux.Vars(r)["colID"]
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	json.NewDecoder(r.Body).Decode(&in)
	if in.Title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	card, err := h.svc.CreateCard(r.Context(), colID, in.Title, in.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get boardID for broadcast — look it up from card's column
	boardID := r.URL.Query().Get("board_id") // hint passed by client
	if boardID != "" {
		h.broadcast(boardID, "card.created", card)
	}
	writeJSON(w, http.StatusCreated, card)
}

func (h *Handler) updateCard(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		BoardID     string `json:"board_id"`
	}
	json.NewDecoder(r.Body).Decode(&in)
	card, err := h.svc.UpdateCard(r.Context(), mux.Vars(r)["cardID"], in.Title, in.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if in.BoardID != "" {
		h.broadcast(in.BoardID, "card.updated", card)
	}
	writeJSON(w, http.StatusOK, card)
}

func (h *Handler) moveCard(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ColumnID string `json:"column_id"`
		Position int    `json:"position"`
		BoardID  string `json:"board_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	// Allow position override via query param for convenience
	if p := r.URL.Query().Get("position"); p != "" {
		pos, _ := strconv.Atoi(p)
		in.Position = pos
	}
	card, err := h.svc.MoveCard(r.Context(), mux.Vars(r)["cardID"], in.ColumnID, in.Position)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if in.BoardID != "" {
		h.broadcast(in.BoardID, "card.moved", card)
	}
	writeJSON(w, http.StatusOK, card)
}

func (h *Handler) deleteCard(w http.ResponseWriter, r *http.Request) {
	cardID := mux.Vars(r)["cardID"]
	boardID := r.URL.Query().Get("board_id")
	if err := h.svc.DeleteCard(r.Context(), cardID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if boardID != "" {
		h.broadcast(boardID, "card.deleted", map[string]string{"id": cardID})
	}
	w.WriteHeader(http.StatusNoContent)
}

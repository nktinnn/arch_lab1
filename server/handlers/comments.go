package handlers

import (
	"encoding/json"
	"net/http"

	"helpdesk/server/middleware"
	"helpdesk/server/models"
)

type CommentHandler struct {
	Comments *models.CommentStore
	Tickets  *models.TicketStore
}

// POST /api/tickets/{id}/comments
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ticketID, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid ticket id", http.StatusBadRequest)
		return
	}

	ticket, err := h.Tickets.GetByID(ticketID)
	if err != nil {
		jsonError(w, "ticket not found", http.StatusNotFound)
		return
	}

	role := middleware.RoleFromCtx(r.Context())
	userID := middleware.UserIDFromCtx(r.Context())

	// user can comment only on own tickets
	if role == models.RoleUser && ticket.AuthorID != userID {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		jsonError(w, "content is required", http.StatusBadRequest)
		return
	}

	comment, err := h.Comments.Create(ticketID, userID, body.Content)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GET /api/tickets/{id}/comments
func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	ticketID, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid ticket id", http.StatusBadRequest)
		return
	}

	ticket, err := h.Tickets.GetByID(ticketID)
	if err != nil {
		jsonError(w, "ticket not found", http.StatusNotFound)
		return
	}

	role := middleware.RoleFromCtx(r.Context())
	userID := middleware.UserIDFromCtx(r.Context())

	if role == models.RoleUser && ticket.AuthorID != userID {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	comments, err := h.Comments.ListByTicket(ticketID)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if comments == nil {
		comments = []*models.Comment{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// DELETE /api/comments/{id} — admin only
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Comments.Delete(id); err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

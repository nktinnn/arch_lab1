package handlers

import (
	"encoding/json"
	"net/http"

	"helpdesk/server/middleware"
	"helpdesk/server/models"
)

type TicketHandler struct {
	Tickets *models.TicketStore
}

// POST /api/tickets
func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string               `json:"title"`
		Description string               `json:"description"`
		Priority    models.TicketPriority `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Title == "" || body.Description == "" {
		jsonError(w, "title and description are required", http.StatusBadRequest)
		return
	}
	if body.Priority == "" {
		body.Priority = models.PriorityMedium
	}

	authorID := middleware.UserIDFromCtx(r.Context())
	ticket, err := h.Tickets.Create(body.Title, body.Description, body.Priority, authorID)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

// GET /api/tickets
// admin/operator see all; user sees own
func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	role := middleware.RoleFromCtx(r.Context())
	userID := middleware.UserIDFromCtx(r.Context())

	var authorFilter *int
	if role == models.RoleUser {
		authorFilter = &userID
	}

	tickets, err := h.Tickets.List(authorFilter)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tickets == nil {
		tickets = []*models.Ticket{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

// GET /api/tickets/{id}
func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	ticket, err := h.Tickets.GetByID(id)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

// PUT /api/tickets/{id}
// admin/operator: full update; user: can only update own open tickets (title/description/priority)
func (h *TicketHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	ticket, err := h.Tickets.GetByID(id)
	if err != nil {
		jsonError(w, "ticket not found", http.StatusNotFound)
		return
	}

	role := middleware.RoleFromCtx(r.Context())
	userID := middleware.UserIDFromCtx(r.Context())

	if role == models.RoleUser {
		if ticket.AuthorID != userID {
			jsonError(w, "forbidden", http.StatusForbidden)
			return
		}
		if ticket.Status != models.StatusOpen {
			jsonError(w, "can only edit open tickets", http.StatusForbidden)
			return
		}
	}

	var body struct {
		Title       string               `json:"title"`
		Description string               `json:"description"`
		Priority    models.TicketPriority `json:"priority"`
		Status      models.TicketStatus  `json:"status"`
		AssignedTo  *int                 `json:"assigned_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}

	// fill defaults from existing ticket
	if body.Title == "" {
		body.Title = ticket.Title
	}
	if body.Description == "" {
		body.Description = ticket.Description
	}
	if body.Priority == "" {
		body.Priority = ticket.Priority
	}
	if body.Status == "" {
		body.Status = ticket.Status
	}

	if role == models.RoleUser {
		// users cannot change status
		body.Status = ticket.Status
	}

	if err := h.Tickets.Update(id, body.Title, body.Description, body.Priority, body.Status); err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if role != models.RoleUser && body.AssignedTo != ticket.AssignedTo {
		h.Tickets.UpdateAssignee(id, body.AssignedTo)
	}

	updated, _ := h.Tickets.GetByID(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DELETE /api/tickets/{id} — admin only
func (h *TicketHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Tickets.Delete(id); err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

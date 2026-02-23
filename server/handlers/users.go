package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"helpdesk/server/models"
)

type UserHandler struct {
	Users *models.UserStore
}

// GET /api/users  — admin only
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.Users.List()
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []*models.User{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// PUT /api/users/{id}/role  — admin only
func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body struct {
		Role models.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Role != models.RoleAdmin && body.Role != models.RoleOperator && body.Role != models.RoleUser {
		jsonError(w, "invalid role", http.StatusBadRequest)
		return
	}

	if err := h.Users.UpdateRole(id, body.Role); err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DELETE /api/users/{id}  — admin only
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Users.Delete(id); err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func pathID(r *http.Request, name string) (int, error) {
	return strconv.Atoi(r.PathValue(name))
}

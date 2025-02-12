package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// createGroup handles POST /groups/
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Verify authentication
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var requestBody struct {
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if requestBody.Name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}
	if len(requestBody.Members) < 2 {
		http.Error(w, "At least 2 members are required", http.StatusBadRequest)
		return
	}
	if len(requestBody.Members) > 50 {
		http.Error(w, "Maximum 50 members allowed", http.StatusBadRequest)
		return
	}

	// Create group with members
	group, err := rt.db.CreateGroup(requestBody.Name, user.ID, requestBody.Members)
	if err != nil {
		http.Error(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	response := struct {
		GroupID string `json:"group_id"`
	}{
		GroupID: group.ID,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// updateGroupName maneja POST /groups/{group_id}
func (rt *_router) updateGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	groupID := ps.ByName("group_id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Verify authentication
	_, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse body
	var requestBody struct {
		NewName string `json:"new_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update name
	err = rt.db.UpdateGroupName(groupID, requestBody.NewName)
	if err != nil {
		http.Error(w, "Failed to update group name: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updateGroupPhoto maneja POST /groups/{group_id}/photo
func (rt *_router) updateGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	groupID := ps.ByName("group_id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Verificar autenticación
	_, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parsear body
	var requestBody struct {
		PhotoURL string `json:"photo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Actualizar foto
	err = rt.db.UpdateGroupPhoto(groupID, requestBody.PhotoURL)
	if err != nil {
		http.Error(w, "Failed to update group photo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// leaveGroup maneja POST /groups/{group_id}/leave
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	groupID := ps.ByName("group_id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Verificar autenticación
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Abandonar grupo
	err = rt.db.LeaveGroup(groupID, user.ID)
	if err != nil {
		http.Error(w, "Failed to leave group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

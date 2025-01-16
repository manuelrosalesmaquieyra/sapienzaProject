package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// createGroup maneja POST /conversations/groups/
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Verificar autenticación
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parsear body
	var requestBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Crear grupo
	group, err := rt.db.CreateGroup(requestBody.Name, user.ID)
	if err != nil {
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// Devolver respuesta
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(group); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// updateGroupName maneja POST /conversations/groups/{group_id}
func (rt *_router) updateGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Actualizar nombre
	err = rt.db.UpdateGroupName(groupID, requestBody.Name)
	if err != nil {
		http.Error(w, "Failed to update group name", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updateGroupPhoto maneja POST /conversations/groups/{group_id}/photo
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

// leaveGroup maneja POST /conversations/groups/{group_id}/leave
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

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Verify authentication
	_, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		http.Error(w, "File must be an image", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("group_%s_%d%s",
		groupID,
		time.Now().UnixNano(),
		filepath.Ext(header.Filename))

	// Save file to uploads/images directory
	filepath := filepath.Join("uploads", "images", filename)
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Create the URL that points to your backend server
	photoURL := fmt.Sprintf("http://localhost:3000/uploads/images/%s", filename)
	if err := rt.db.UpdateGroupPhoto(groupID, photoURL); err != nil {
		http.Error(w, "Failed to update photo in database", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := struct {
		PhotoURL string `json:"photo_url"`
	}{
		PhotoURL: photoURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

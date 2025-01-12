package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName maneja PUT /users/{username}
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check if the request method is PUT
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get username from URL params
	username := ps.ByName("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Validate username format
	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernamePattern.MatchString(username) || len(username) < 3 || len(username) > 16 {
		http.Error(w, "Invalid username format", http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var requestBody struct {
		NewName string `json:"new_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate new_name format
	if !usernamePattern.MatchString(requestBody.NewName) {
		http.Error(w, "Invalid new_name format", http.StatusBadRequest)
		return
	}

	// Validate new_name length
	if len(requestBody.NewName) < 3 || len(requestBody.NewName) > 16 {
		http.Error(w, "Invalid new_name length", http.StatusBadRequest)
		return
	}

	// Get and validate auth token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return
	}
	token := authHeader[7:]

	// Verify user is authorized to change this username
	user, err := rt.db.GetUserByToken(token)
	if err != nil || user.Username != username {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update username in database
	if err := rt.db.UpdateUsername(user.ID, requestBody.NewName); err != nil {
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := struct {
		Message  string `json:"message"`
		Username string `json:"username"`
	}{
		Message:  "Username successfully updated",
		Username: requestBody.NewName,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// setMyPhoto maneja POST /users/{username}/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	username := ps.ByName("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var requestBody struct {
		PhotoURL string `json:"photo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate photo URL
	if requestBody.PhotoURL == "" {
		http.Error(w, "Photo URL is required", http.StatusBadRequest)
		return
	}

	// Get user from token
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update photo
	if err := rt.db.UpdateUserPhoto(user.ID, requestBody.PhotoURL); err != nil {
		http.Error(w, "Failed to update photo", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := struct {
		PhotoURL string `json:"photo_url"`
	}{
		PhotoURL: requestBody.PhotoURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getUserConversations maneja GET /users/{username}/conversations
func (rt *_router) getUserConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	username := ps.ByName("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Get user from token
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get conversations
	conversations, err := rt.db.GetUserConversations(user.ID)
	if err != nil {
		http.Error(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

	// Return conversations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// Función auxiliar para obtener usuario desde token
func (rt *_router) getUserFromToken(r *http.Request) (*database.User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		return nil, errors.New("invalid authorization header")
	}
	token := authHeader[7:]

	return rt.db.GetUserByToken(token)
}

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
		// Check for specific error messages
		switch {
		case err.Error() == "new username is the same as current username":
			http.Error(w, err.Error(), http.StatusBadRequest)
		case err.Error() == fmt.Sprintf("username '%s' is already taken, please choose a different one", requestBody.NewName):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			log.Printf("Error updating username: %v", err)
			http.Error(w, "Failed to update username", http.StatusInternalServerError)
		}
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// setMyPhoto maneja POST /users/{username}/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	filename := fmt.Sprintf("%s_%d%s",
		user.ID,
		time.Now().UnixNano(),
		filepath.Ext(header.Filename))

	// Ensure the uploads/images directory exists
	uploadDir := filepath.Join("uploads", "images")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Failed to ensure upload directory exists", http.StatusInternalServerError)
		return
	}

	// Save file to uploads/images directory
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
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
	if err := rt.db.UpdateUserPhoto(user.ID, photoURL); err != nil {
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

// getUserConversations maneja GET /users/{username}/conversations
// func (rt *_router) getUserConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	log.Printf("Handler called with user ID: %v", ps.ByName("username"))
// 	username := ps.ByName("username")
// 	if username == "" {
// 		http.Error(w, "Username is required", http.StatusBadRequest)
// 		return
// 	}

// 	// Get user from token
// 	user, err := rt.getUserFromToken(r)
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Get conversations
// 	conversations, err := rt.db.GetUserConversations(user.ID)
// 	if err != nil {
// 		http.Error(w, "Failed to get conversations", http.StatusInternalServerError)
// 		return
// 	}

// 	// Return conversations
// 	w.Header().Set("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(conversations); err != nil {
// 		http.Error(w, "Error encoding response", http.StatusInternalServerError)
// 		return
// 	}
// }

// Función auxiliar para obtener usuario desde token
func (rt *_router) getUserFromToken(r *http.Request) (*database.User, error) {
	authHeader := r.Header.Get("Authorization")
	log.Printf("Auth header received: %s", authHeader)

	if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		log.Printf("Invalid auth header format")
		return nil, errors.New("invalid authorization header")
	}

	token := authHeader[7:]
	log.Printf("Looking for token: %s", token)

	user, err := rt.db.GetUserByToken(token)
	if err != nil {
		log.Printf("Error getting user by token: %v", err)
	}
	return user, err
}

func (rt *_router) getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// Only return data if the requested username matches the authenticated user
	if username != user.Username {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return user data
	response := struct {
		Username string `json:"username"`
		PhotoURL string `json:"photo_url"`
	}{
		Username: user.Username,
		PhotoURL: user.PhotoURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) checkUserExists(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	username := ps.ByName("username")

	// Check if user exists in the database
	exists := rt.db.HasUser(username)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Return the existence status
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists": exists,
	})
}

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// deleteMessage maneja DELETE /messages/{messageId}
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageID := ps.ByName("messageId")
	conversationID := ps.ByName("conversationId")
	if messageID == "" || conversationID == "" {
		http.Error(w, "Message ID and Conversation ID are required", http.StatusBadRequest)
		return
	}

	// Verificar autenticación
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verificar que el usuario es el remitente del mensaje
	message, err := rt.db.GetMessageByID(messageID)
	if err != nil || message.Sender != user.Username {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Eliminar mensaje
	err = rt.db.DeleteMessage(messageID)
	if err != nil {
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// forwardMessage maneja POST conversations/{conversationId}/messages/{messageId}
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	targetConversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")

	// Get authenticated user
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Forward the message using the updated database method
	newMessage, err := rt.db.ForwardMessage(messageID, targetConversationID, user.Username)
	if err != nil {
		http.Error(w, "Failed to forward message", http.StatusInternalServerError)
		return
	}

	// Return the new message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newMessage); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) replyToMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")

	// Get authenticated user
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create reply message
	newMessageID, err := rt.db.CreateReplyMessage(conversationID, user.Username, req.Content, messageID)
	if err != nil {
		http.Error(w, "Failed to create reply", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message_id": newMessageID,
	})
}

func (rt *_router) sendImageMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationID := ps.ByName("conversationId")

	// Get authenticated user
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set max upload size to 20MB
	maxSize := int64(20 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	// Parse the multipart form with increased size limit
	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, "Failed to parse form. Make sure the image is under 20MB", http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Verify file type
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		http.Error(w, "File must be an image", http.StatusBadRequest)
		return
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("uploads/images", 0755); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s-%s", uuid.New().String(), header.Filename)
	filepath := path.Join("uploads/images", filename)

	// Save file
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	// Create message with image URL
	imageURL := fmt.Sprintf("/uploads/images/%s", filename)
	newMessageID, err := rt.db.CreateImageMessage(conversationID, user.Username, imageURL)
	if err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	// Return the full URL in the response
	fullImageURL := fmt.Sprintf("http://localhost:3000%s", imageURL)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message_id": newMessageID,
		"image_url":  fullImageURL,
	})
}

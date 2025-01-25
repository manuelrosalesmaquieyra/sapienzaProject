package api

import (
	"encoding/json"
	"log"
	"net/http"

	//"sapienzaProject/service/database"

	"github.com/julienschmidt/httprouter"
)

// getConversationMessages maneja GET /conversations/{conversationId}/messages
func (rt *_router) getConversationMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get conversation ID from URL
	conversationId := ps.ByName("conversationId")
	if conversationId == "" {
		http.Error(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

	// Get authenticated user
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify user is part of the conversation
	isParticipant, err := rt.db.IsUserInConversation(user.Username, conversationId)
	if err != nil {
		http.Error(w, "Error checking conversation access", http.StatusInternalServerError)
		return
	}
	if !isParticipant {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get messages
	messages, err := rt.db.GetConversationMessages(conversationId)
	if err != nil {
		log.Printf("Error getting messages: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	// Return messages
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	}); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// sendMessage maneja POST /conversations/{conversationId}/messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get conversation ID from URL
	conversationId := ps.ByName("conversationId")
	if conversationId == "" {
		http.Error(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

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

	// Create message
	messageId, err := rt.db.CreateMessage(conversationId, user.Username, req.Content)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message_id": messageId,
	}); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// createConversation handles POST /conversations
func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get authenticated user
	user, err := rt.getUserFromToken(r)
	if err != nil {
		log.Printf("Auth error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("Authenticated user: %s", user.Username)

	// Parse request body
	var req struct {
		Participants []string `json:"participants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Request decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("Received participants from request: %v", req.Participants)

	// Ensure we have exactly one other participant
	if len(req.Participants) != 1 {
		log.Printf("Invalid number of participants: %d", len(req.Participants))
		http.Error(w, "Must specify exactly one participant", http.StatusBadRequest)
		return
	}

	// Create participants array with both users
	participants := []string{user.Username, req.Participants[0]}
	log.Printf("Final participants list: %v", participants)

	// Create conversation
	conversationID, err := rt.db.CreateConversation(participants)
	if err != nil {
		log.Printf("Create conversation error: %v", err)
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}
	log.Printf("Created conversation: %s", conversationID)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"conversation_id": conversationID,
	}); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
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

	// Verify the user is requesting their own conversations
	if user.Username != username {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get conversations using username
	conversations, err := rt.db.GetUserConversations(user.Username)
	if err != nil {
		log.Printf("Error getting conversations: %v", err)
		http.Error(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

	// Return conversations
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conversations); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get conversation ID from params
	conversationId := ps.ByName("conversationId")

	// Get user from token
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is in conversation
	isParticipant, err := rt.db.IsUserInConversation(user.Username, conversationId)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isParticipant {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get messages
	messages, err := rt.db.GetConversationMessages(conversationId)
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	}); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

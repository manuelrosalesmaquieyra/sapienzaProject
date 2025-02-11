package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// addReaction maneja POST /conversations/{conversationId}/messages/{messageId}/reactions
func (rt *_router) addReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageId := ps.ByName("messageId")
	//conversationId := ps.ByName("conversationId")

	if messageId == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	// Verificar autenticación
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parsear body
	var requestBody struct {
		Reaction string `json:"reaction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Adding reaction - MessageID: %s, UserID: %s, Reaction: %s", messageId, user.ID, requestBody.Reaction)

	// Añadir reacción
	err = rt.db.AddReaction(messageId, user.ID, requestBody.Reaction)
	if err != nil {
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// removeReaction maneja DELETE /conversations/{conversationId}/messages/{messageId}/reactions
func (rt *_router) removeReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageId := ps.ByName("messageId")
	conversationId := ps.ByName("conversationId")

	if messageId == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	// Verificar autenticación
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("Removing reaction - MessageID: %s, UserID: %s, ConversationID: %s", messageId, user.ID, conversationId)

	// Eliminar reacción
	err = rt.db.RemoveReaction(messageId, user.ID)
	if err != nil {
		log.Printf("Error removing reaction: %v", err)
		http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package api

import (
	"encoding/json"
	"net/http"

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

	// If messageID is provided, it's a forward operation
	if messageID != "" {
		// Get authenticated user
		user, err := rt.getUserFromToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get original message
		originalMessage, err := rt.db.GetMessageByID(messageID)
		if err != nil {
			http.Error(w, "Message not found", http.StatusNotFound)
			return
		}

		// Create forwarded message with current user as sender
		newMessageID, err := rt.db.CreateMessage(
			targetConversationID,
			user.Username, // Set current user as sender
			originalMessage.Content,
		)
		if err != nil {
			http.Error(w, "Failed to forward message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message_id": newMessageID,
		})
		return
	}

	// If no messageID, handle as regular new message
	// ... your existing new message creation code ...
}

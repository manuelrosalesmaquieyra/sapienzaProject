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
	if err != nil || message.Sender != user.ID {
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

	// Reenviar mensaje
	newMessage, err := rt.db.ForwardMessage(messageID, conversationID, user.ID)
	if err != nil {
		http.Error(w, "Failed to forward message", http.StatusInternalServerError)
		return
	}

	// Devolver respuesta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newMessage)
}

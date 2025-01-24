package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// addReaction maneja POST /conversations/{conversation_id}/messages/{message_id}/reactions
func (rt *_router) addReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageID := ps.ByName("messageId")
	if messageID == "" {
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

	// Añadir reacción
	err = rt.db.AddReaction(messageID, user.ID, requestBody.Reaction)
	if err != nil {
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// removeReaction maneja DELETE /conversations/{conversation_id}/messages/{message_id}/reactions
func (rt *_router) removeReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageID := ps.ByName("messageId")
	if messageID == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	// Verificar autenticación
	user, err := rt.getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Eliminar reacción
	err = rt.db.RemoveReaction(messageID, user.ID)
	if err != nil {
		http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// loginRequest representa la estructura del body de la petición
type loginRequest struct {
	Name string `json:"name"`
}

// loginResponse representa la estructura de la respuesta
type loginResponse struct {
	Username   string `json:"username"`
	Identifier string `json:"session_id"`
}

// doLogin maneja POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Decodificar el body
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Crear sesión
	session, err := rt.db.CreateSession(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create response
	response := loginResponse{
		Username:   session.Username,
		Identifier: session.Identifier,
	}

	// Enviar respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

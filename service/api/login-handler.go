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

	// Enviar respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/opi-es/web-todo/pkg/auth"
)

type AuthRequest struct {
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	envPass := os.Getenv("TODO_PASSWORD")
	if envPass == "" {
		writeJSONError(w, "Authentication not configured", http.StatusInternalServerError)
		return
	}

	if req.Password != envPass {
		writeJSONError(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		writeJSONError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{Token: token})
}

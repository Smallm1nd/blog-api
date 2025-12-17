package handlers

import (
	"blog_api/internal/models"
	"blog_api/internal/server"
	"encoding/json"
	"net/http"
)

// post user
func RegisterUser(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Methon not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "all fields required", http.StatusBadRequest)
			return
		}

		var userID int
		err := s.DB.QueryRow(
			"INSERT INTO users (username, email, password) VALUES($1, $2, $3) RETURNING id",
			req.Username, req.Email, req.Password,
		).Scan(&userID)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		resp := models.User{
			ID:       userID,
			Username: req.Username,
			Email:    req.Email,
		}

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

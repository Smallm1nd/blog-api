package handlers

import (
	"blog_api/internal/models"
	"blog_api/internal/server"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func LikePost(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		postID, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/posts/"), "/likes"))
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		var UserID struct {
			UserID int `json:"user_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&UserID)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if UserID.UserID == 0 {
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}

		var like models.Like
		err = s.DB.QueryRow("INSERT INTO likes (post_id, user_id) VALUES($1, $2) RETURNING id, created_at",
			postID, UserID.UserID).Scan(&like.ID, &like.CreatedAt)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				http.Error(w, "Already liked", http.StatusConflict)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		like.UserID = UserID.UserID
		like.PostID = postID

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(like)
	}
}

func UnlikePost(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		postID, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/posts/"), "/likes"))
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}
		var UserID struct {
			UserID int `json:"user_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&UserID)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if UserID.UserID == 0 {
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}

		result, err := s.DB.Exec("DELETE FROM likes WHERE post_id = $1 AND user_id = $2",
			postID, UserID.UserID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		rows, err := result.RowsAffected()
		if err != nil {
			http.Error(w, "Error checking result", http.StatusInternalServerError)
			return
		}
		if rows == 0 {
			http.Error(w, "Like not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Unlike successfully",
		})
	}
}

func GetLikes(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		postID, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/posts/"), "/likes"))
		if err != nil {
			http.Error(w, "Not found page", http.StatusNotFound)
			return
		}

		var countLike int
		err = s.DB.QueryRow("SELECT COUNT(id) FROM likes WHERE post_id = $1",
			postID,
		).Scan(&countLike)

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{
			"likes": countLike,
		})
	}
}

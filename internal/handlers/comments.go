package handlers

import (
	"blog_api/internal/models"
	"blog_api/internal/server"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func CreateComment(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PostID  int    `json:"post_id"`
			UserID  int    `json:"user_id"`
			Content string `json:"content"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Content == "" || req.PostID == 0 || req.UserID == 0 {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		var resp models.Comment
		err = s.DB.QueryRow("INSERT INTO comments (user_id, post_id, content) VALUES($1, $2, $3) RETURNING id, created_at",
			req.UserID, req.PostID, req.Content).Scan(&resp.ID, &resp.CreatedAt)

		if err != nil {
			http.Error(w, "Failed to create comment", http.StatusInternalServerError)
			return
		}

		resp.Content = req.Content
		resp.PostID = req.PostID
		resp.UserID = req.UserID

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)

	}
}

func GetComment(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var commentID int
		commentID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/comments/"))
		if err != nil {
			http.Error(w, "Not found page", http.StatusNotFound)
			return
		}

		var resp models.Comment
		err = s.DB.QueryRow("SELECT id, user_id, post_id, content, created_at FROM comments WHERE id = $1",
			commentID).Scan(&resp.ID, &resp.UserID, &resp.PostID, &resp.Content, &resp.CreatedAt)
		if err == sql.ErrNoRows {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func DeleteComment(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) != 3 {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		commentID, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		result, err := s.DB.Exec("DELETE FROM comments WHERE id = $1", commentID)
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
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Comment deleted successfully",
		})
	}
}

func GetAllPostsComments(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var postID int
		postID, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/posts/"), "/comments"))
		if err != nil {
			http.Error(w, "Not found page", http.StatusNotFound)
			return
		}

		rows, err := s.DB.Query(
			"SELECT id, user_id, post_id, content, created_at FROM comments WHERE post_id = $1 ORDER BY created_at DESC",
			postID)

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		comments := []models.Comment{}
		for rows.Next() {
			var comment models.Comment
			err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.Content, &comment.CreatedAt)
			if err != nil {
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			comments = append(comments, comment)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, "Rows error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(comments)
	}
}

func CommentsSpecialHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			GetComment(s)(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			DeleteComment(s)(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

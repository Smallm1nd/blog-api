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

func CreatePost(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			UserID  int    `json:"user_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Inviled JSON", http.StatusBadRequest)
			return
		}

		if req.Content == "" || req.Title == "" || req.UserID == 0 {
			http.Error(w, "all fields required", http.StatusBadRequest)
			return
		}

		var resp models.Post
		err := s.DB.QueryRow("INSERT INTO posts (title, content, user_id) VALUES($1, $2, $3) RETURNING id, created_at",
			req.Title, req.Content, req.UserID,
		).Scan(&resp.ID, &resp.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		resp.Title = req.Title
		resp.Content = req.Content
		resp.UserID = req.UserID

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

func GetPost(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/posts/")
		postID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Not found page", http.StatusNotFound)
			return
		}

		var respPost models.Post
		err = s.DB.QueryRow("SELECT id, title, content, user_id, created_at FROM posts WHERE id = $1",
			postID,
		).Scan(&respPost.ID, &respPost.Title, &respPost.Content, &respPost.UserID, &respPost.CreatedAt)

		if err == sql.ErrNoRows {
			http.Error(w, "Post not Found", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respPost)
	}
}

func GetAllPosts(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var pageStr string = r.URL.Query().Get("page")
		var limitStr string = r.URL.Query().Get("limit")

		var page int = 1
		var limit int = 10

		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		var offset int = (page - 1) * limit

		rows, err := s.DB.Query(`
		SELECT * FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
		`, limit, offset)
		if err != nil {
			http.Error(w, "invilad rows", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []models.Post
		for rows.Next() {
			var post models.Post
			err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
			if err != nil {
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			posts = append(posts, post)
		}
		if err = rows.Err(); err != nil {
			http.Error(w, "Rows error", http.StatusInternalServerError)
			return
		}

		var total int
		s.DB.QueryRow("SELECT COUNT(*) FROM posts").Scan(&total)

		resp := map[string]interface{}{
			"posts": posts,
			"page":  page,
			"limit": limit,
			"total": total,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func PostsSpecialHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			CreatePost(s)(w, r)
			return
		}

		if r.Method == http.MethodGet {
			GetAllPosts(s)(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

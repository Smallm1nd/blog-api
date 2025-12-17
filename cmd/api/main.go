package main

import (
	"blog_api/internal/database"
	"blog_api/internal/handlers"
	"blog_api/internal/server"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	db, err := database.Connect("localhost", "5432", "postgres", "killer", "blog_db")
	if err != nil {
		log.Fatal("conect to database failed ", err)
	}
	defer db.Close()
	fmt.Println("✅ Connected to database!")

	srv := server.NewServer(db)

	http.HandleFunc("/users/register", handlers.RegisterUser(srv))

	http.HandleFunc("/comments/", handlers.CommentsSpecialHandler(srv))
	http.HandleFunc("/comments", handlers.CreateComment(srv))

	http.HandleFunc("/posts", handlers.PostsSpecialHandler(srv))

	http.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/comments") {
			handlers.GetAllPostsComments(srv)(w, r)
		} else {
			handlers.GetPost(srv)(w, r)
		}
	})

	fmt.Println("🚀 Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

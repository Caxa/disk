package main

import (
	"fmt"
	"net/http"

	"disk/handlers"

	_ "github.com/lib/pq"
)

func main() {

	handlers.OpenDatabase()
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/main", handlers.Index)
	http.HandleFunc("/upload", handlers.UploadFile)
	http.HandleFunc("/download/", handlers.DownloadFile)
	http.HandleFunc("/delete", handlers.DeleteFile)
	http.HandleFunc("/sent_files", handlers.Sent)
	fmt.Println("Server is running on http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}

package main

import (
	"log"
	"net/http"

	httptransport "mentorforge-ai/backend-go/internal/transport/http"
)

func main() {
	handler := httptransport.NewHandler()

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler.Routes(),
	}

	log.Println("mentorforge-ai is listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

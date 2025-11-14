package main

import (
	"infraexplain/api/router"
	"log"
	"net/http"
)

func main() {
	mux := router.SetupRoutes()

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

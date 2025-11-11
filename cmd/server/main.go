package main

import (
	"log"

	"lam-phuong-api/internal/book"
	"lam-phuong-api/internal/location"
	"lam-phuong-api/internal/server"
)

func main() {
	bookSeed := []book.Book{
		{ID: "1", Title: "The Go Programming Language", Author: "Alan A. A. Donovan"},
		{ID: "2", Title: "Introducing Go", Author: "Caleb Doxsey"},
	}

	locationSeed := []location.Location{
		{ID: "1", Name: "Main Library", Address: "123 Main St", City: "Go City"},
		{ID: "2", Name: "West Branch", Address: "456 Elm St", City: "Go City"},
	}

	bookRepo := book.NewInMemoryRepository(bookSeed)
	bookHandler := book.NewHandler(bookRepo)

	locationRepo := location.NewInMemoryRepository(locationSeed)
	locationHandler := location.NewHandler(locationRepo)

	router := server.NewRouter(bookHandler, locationHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

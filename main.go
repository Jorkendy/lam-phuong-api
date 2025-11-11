package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Book represents a simple in-memory record for demo purposes.
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var books = []Book{
	{ID: "1", Title: "The Go Programming Language", Author: "Alan A. A. Donovan"},
	{ID: "2", Title: "Introducing Go", Author: "Caleb Doxsey"},
}

func main() {
	router := gin.Default()

	api := router.Group("/api")
	{
		api.GET("/ping", pingHandler)
		api.GET("/books", listBooksHandler)
		api.GET("/books/:id", getBookHandler)
		api.POST("/books", createBookHandler)
		api.PUT("/books/:id", updateBookHandler)
		api.DELETE("/books/:id", deleteBookHandler)
	}

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func listBooksHandler(c *gin.Context) {
	c.JSON(http.StatusOK, books)
}

func getBookHandler(c *gin.Context) {
	id := c.Param("id")
	for _, book := range books {
		if book.ID == id {
			c.JSON(http.StatusOK, book)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
}

func createBookHandler(c *gin.Context) {
	var book Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Basic ID assignment for demo purposes.
	book.ID = generateNextID()
	books = append(books, book)

	c.JSON(http.StatusCreated, book)
}

func updateBookHandler(c *gin.Context) {
	id := c.Param("id")
	var payload Book
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, book := range books {
		if book.ID == id {
			books[i].Title = payload.Title
			books[i].Author = payload.Author
			c.JSON(http.StatusOK, books[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
}

func deleteBookHandler(c *gin.Context) {
	id := c.Param("id")

	for i, book := range books {
		if book.ID == id {
			books = append(books[:i], books[i+1:]...)
			c.Status(http.StatusNoContent)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
}

func generateNextID() string {
	next := len(books) + 1
	return fmt.Sprintf("%d", next)
}

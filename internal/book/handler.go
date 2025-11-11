package book

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP handlers for the book resource.
type Handler struct {
	repo Repository
}

// NewHandler creates a handler with the provided repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes attaches book routes to the supplied router group.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/books", h.ListBooks)
	router.GET("/books/:id", h.GetBook)
	router.POST("/books", h.CreateBook)
	router.PUT("/books/:id", h.UpdateBook)
	router.DELETE("/books/:id", h.DeleteBook)
}

func (h *Handler) ListBooks(c *gin.Context) {
	c.JSON(http.StatusOK, h.repo.List())
}

func (h *Handler) GetBook(c *gin.Context) {
	id := c.Param("id")
	book, ok := h.repo.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *Handler) CreateBook(c *gin.Context) {
	var payload bookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book := Book{
		Title:  payload.Title,
		Author: payload.Author,
	}

	created, err := h.repo.Create(book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) UpdateBook(c *gin.Context) {
	id := c.Param("id")

	var payload bookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book := Book{
		Title:  payload.Title,
		Author: payload.Author,
	}

	updated, ok := h.repo.Update(id, book)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) DeleteBook(c *gin.Context) {
	id := c.Param("id")
	if ok := h.repo.Delete(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

type bookPayload struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
}

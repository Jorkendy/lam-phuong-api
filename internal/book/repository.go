package book

import (
	"sort"
	"strconv"
	"sync"
)

// Repository defines behavior for storing and retrieving books.
type Repository interface {
	List() []Book
	Get(id string) (Book, bool)
	Create(book Book) (Book, error)
	Update(id string, book Book) (Book, bool)
	Delete(id string) bool
}

// InMemoryRepository stores books in memory and is safe for concurrent access.
type InMemoryRepository struct {
	mu     sync.RWMutex
	books  map[string]Book
	nextID int
}

// NewInMemoryRepository creates an in-memory repository seeded with optional data.
func NewInMemoryRepository(seed []Book) *InMemoryRepository {
	repo := &InMemoryRepository{
		books:  make(map[string]Book),
		nextID: 1,
	}

	maxID := 0
	for _, b := range seed {
		repo.books[b.ID] = b
		if id, err := strconv.Atoi(b.ID); err == nil && id > maxID {
			maxID = id
		}
	}
	repo.nextID = maxID + 1

	return repo
}

// List returns all books sorted by ID.
func (r *InMemoryRepository) List() []Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	books := make([]Book, 0, len(r.books))
	for _, book := range r.books {
		books = append(books, book)
	}

	sort.Slice(books, func(i, j int) bool {
		return books[i].ID < books[j].ID
	})

	return books
}

// Get retrieves a book by ID.
func (r *InMemoryRepository) Get(id string) (Book, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book, ok := r.books[id]
	return book, ok
}

// Create adds a new book and automatically assigns an ID.
func (r *InMemoryRepository) Create(book Book) (Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	book.ID = strconv.Itoa(r.nextID)
	r.nextID++
	r.books[book.ID] = book

	return book, nil
}

// Update modifies an existing book record.
func (r *InMemoryRepository) Update(id string, book Book) (Book, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[id]; !exists {
		return Book{}, false
	}

	book.ID = id
	r.books[id] = book
	return book, true
}

// Delete removes a book by ID.
func (r *InMemoryRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[id]; !exists {
		return false
	}

	delete(r.books, id)
	return true
}

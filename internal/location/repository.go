package location

import (
	"sort"
	"strconv"
	"sync"
)

// Repository defines behavior for storing and retrieving locations.
type Repository interface {
	List() []Location
	Get(id string) (Location, bool)
	Create(location Location) (Location, error)
	Update(id string, location Location) (Location, bool)
	Delete(id string) bool
}

// InMemoryRepository stores locations in memory and is safe for concurrent access.
type InMemoryRepository struct {
	mu     sync.RWMutex
	data   map[string]Location
	nextID int
}

// NewInMemoryRepository creates an in-memory repository seeded with optional data.
func NewInMemoryRepository(seed []Location) *InMemoryRepository {
	repo := &InMemoryRepository{
		data:   make(map[string]Location),
		nextID: 1,
	}

	maxID := 0
	for _, l := range seed {
		repo.data[l.ID] = l
		if id, err := strconv.Atoi(l.ID); err == nil && id > maxID {
			maxID = id
		}
	}
	repo.nextID = maxID + 1

	return repo
}

// List returns all locations sorted by ID.
func (r *InMemoryRepository) List() []Location {
	r.mu.RLock()
	defer r.mu.RUnlock()

	locations := make([]Location, 0, len(r.data))
	for _, location := range r.data {
		locations = append(locations, location)
	}

	sort.Slice(locations, func(i, j int) bool {
		return locations[i].ID < locations[j].ID
	})

	return locations
}

// Get retrieves a location by ID.
func (r *InMemoryRepository) Get(id string) (Location, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	location, ok := r.data[id]
	return location, ok
}

// Create adds a new location and automatically assigns an ID.
func (r *InMemoryRepository) Create(location Location) (Location, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	location.ID = strconv.Itoa(r.nextID)
	r.nextID++
	r.data[location.ID] = location

	return location, nil
}

// Update modifies an existing location record.
func (r *InMemoryRepository) Update(id string, location Location) (Location, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return Location{}, false
	}

	location.ID = id
	r.data[id] = location
	return location, true
}

// Delete removes a location by ID.
func (r *InMemoryRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return false
	}

	delete(r.data, id)
	return true
}

package location

// Location represents a physical place served by the API.
type Location struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
}

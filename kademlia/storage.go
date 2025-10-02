package kademlia

// Storage holds a map of KademliaID (hashes) to stored values.
type Storage struct {
	data map[KademliaID]string
}

// NewStorage initializes and returns a new instance of Storage.
func NewStorage() *Storage {
	return &Storage{
		data: make(map[KademliaID]string),
	}
}

// StoreValue stores a value associated with a given KademliaID.
func (s *Storage) StoreValue(id KademliaID, value string) {
	s.data[id] = value
	// log.Printf("(File: Storage, Function: StoreValue) Store value: %s Store id: %s", value, id.String())
}

// GetValue retrieves a value associated with a given KademliaID.
// Returns the value and a boolean indicating whether the value was found.
func (s *Storage) GetValue(id KademliaID) (string, bool) {
	value, exists := s.data[id]
	return value, exists
}

// DeleteValue removes a value associated with a given KademliaID from storage.
func (s *Storage) DeleteValue(id KademliaID) {
	delete(s.data, id)
}

// HasValue checks if a value exists for a given KademliaID.
func (s *Storage) HasValue(id KademliaID) bool {
	_, exists := s.data[id]
	return exists
}
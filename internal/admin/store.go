package admin

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Instance represents a remote Kurokku clock connected via Redis.
type Instance struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Store persists instances to a local JSON file.
type Store struct {
	path      string
	mu        sync.Mutex
	instances []Instance
}

// NewStore creates a Store backed by the given file path.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Load reads instances from the backing file.
// If the file doesn't exist, it starts with an empty list.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		s.instances = nil
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading store: %w", err)
	}
	var instances []Instance
	if err := json.Unmarshal(data, &instances); err != nil {
		return fmt.Errorf("parsing store: %w", err)
	}
	s.instances = instances
	return nil
}

// Save writes the current instances to the backing file.
func (s *Store) save() error {
	data, err := json.MarshalIndent(s.instances, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling store: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}

// List returns all instances.
func (s *Store) List() []Instance {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Instance, len(s.instances))
	copy(out, s.instances)
	return out
}

// Get returns an instance by ID, or nil if not found.
func (s *Store) Get(id string) *Instance {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, inst := range s.instances {
		if inst.ID == id {
			cp := inst
			return &cp
		}
	}
	return nil
}

// Add appends an instance and flushes to disk.
func (s *Store) Add(inst Instance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances = append(s.instances, inst)
	return s.save()
}

// Update replaces an instance by ID and flushes to disk.
func (s *Store) Update(inst Instance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.instances {
		if existing.ID == inst.ID {
			s.instances[i] = inst
			return s.save()
		}
	}
	return fmt.Errorf("instance %q not found", inst.ID)
}

// Remove deletes an instance by ID and flushes to disk.
func (s *Store) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, inst := range s.instances {
		if inst.ID == id {
			s.instances = append(s.instances[:i], s.instances[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("instance %q not found", id)
}

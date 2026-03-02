package admin

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	return NewStore(filepath.Join(dir, "instances.json"))
}

func TestStore_Load_MissingFile(t *testing.T) {
	s := tempStore(t)
	if err := s.Load(); err != nil {
		t.Fatalf("Load with missing file should not error: %v", err)
	}
	if got := s.List(); len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}

func TestStore_AddAndList(t *testing.T) {
	s := tempStore(t)
	inst := Instance{ID: "abc", Name: "Test", Host: "localhost", Port: 6379}
	if err := s.Add(inst); err != nil {
		t.Fatal(err)
	}
	list := s.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(list))
	}
	if list[0].ID != "abc" {
		t.Errorf("ID = %q, want %q", list[0].ID, "abc")
	}
	if list[0].Name != "Test" {
		t.Errorf("Name = %q, want %q", list[0].Name, "Test")
	}
}

func TestStore_Get_Found(t *testing.T) {
	s := tempStore(t)
	s.Add(Instance{ID: "x1", Name: "One", Host: "h1", Port: 1})
	s.Add(Instance{ID: "x2", Name: "Two", Host: "h2", Port: 2})

	got := s.Get("x2")
	if got == nil {
		t.Fatal("expected to find instance x2")
	}
	if got.Name != "Two" {
		t.Errorf("Name = %q, want %q", got.Name, "Two")
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	s := tempStore(t)
	if got := s.Get("nonexistent"); got != nil {
		t.Errorf("expected nil for missing instance, got %+v", got)
	}
}

func TestStore_Update(t *testing.T) {
	s := tempStore(t)
	s.Add(Instance{ID: "u1", Name: "Before", Host: "h", Port: 1})

	err := s.Update(Instance{ID: "u1", Name: "After", Host: "h2", Port: 2})
	if err != nil {
		t.Fatal(err)
	}
	got := s.Get("u1")
	if got.Name != "After" {
		t.Errorf("Name = %q, want %q", got.Name, "After")
	}
	if got.Host != "h2" {
		t.Errorf("Host = %q, want %q", got.Host, "h2")
	}
}

func TestStore_Update_NotFound(t *testing.T) {
	s := tempStore(t)
	err := s.Update(Instance{ID: "missing"})
	if err == nil {
		t.Error("expected error updating nonexistent instance")
	}
}

func TestStore_Remove(t *testing.T) {
	s := tempStore(t)
	s.Add(Instance{ID: "r1", Name: "One"})
	s.Add(Instance{ID: "r2", Name: "Two"})

	if err := s.Remove("r1"); err != nil {
		t.Fatal(err)
	}
	if len(s.List()) != 1 {
		t.Fatalf("expected 1 instance after remove, got %d", len(s.List()))
	}
	if s.Get("r1") != nil {
		t.Error("expected r1 to be removed")
	}
}

func TestStore_Remove_NotFound(t *testing.T) {
	s := tempStore(t)
	err := s.Remove("missing")
	if err == nil {
		t.Error("expected error removing nonexistent instance")
	}
}

func TestStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	// Write data with one store
	s1 := NewStore(path)
	s1.Add(Instance{ID: "p1", Name: "Persist", Host: "h", Port: 1})

	// Read data with a fresh store
	s2 := NewStore(path)
	if err := s2.Load(); err != nil {
		t.Fatal(err)
	}
	list := s2.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 persisted instance, got %d", len(list))
	}
	if list[0].ID != "p1" {
		t.Errorf("ID = %q, want %q", list[0].ID, "p1")
	}
}

func TestStore_Load_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte(`not json`), 0644)

	s := NewStore(path)
	err := s.Load()
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestStore_ListReturnsCopy(t *testing.T) {
	s := tempStore(t)
	s.Add(Instance{ID: "c1", Name: "Original"})

	list := s.List()
	list[0].Name = "Modified"

	// Original should be unchanged
	got := s.Get("c1")
	if got.Name != "Original" {
		t.Errorf("List should return a copy; Name changed to %q", got.Name)
	}
}

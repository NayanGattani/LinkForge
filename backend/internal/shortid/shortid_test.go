package shortid

import (
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"zero", 0, "0"},
		{"one", 1, "1"},
		{"base-1", 61, "z"},
		{"base", 62, "10"},
		{"large", 987654321, "14q60P"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.input)
			if got != tt.expected {
				t.Errorf("Encode(%d) = %s; want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGenerateUnique(t *testing.T) {
	err := Init(1)
	if err != nil {
		t.Fatalf("Failed to init node: %v", err)
	}

	// Generate 1000 IDs and ensure none are duplicates
	seen := make(map[uint64]bool)
	var prev uint64
	for i := 0; i < 1000; i++ {
		id := Generate()
		
		// Check uniqueness
		if seen[id] {
			t.Fatalf("Duplicate ID generated: %d", id)
		}
		seen[id] = true
		
		// Check ordering (Snowflake IDs are time-ordered)
		if i > 0 && id <= prev {
			t.Fatalf("IDs are not strictly increasing: current %d, previous %d", id, prev)
		}
		prev = id
	}
}

func TestGenerateBase62(t *testing.T) {
	err := Init(1)
	if err != nil {
		t.Fatalf("Failed to init node: %v", err)
	}

	res := GenerateBase62()
	if res == "" {
		t.Fatal("Expected non-empty base62 string")
	}
}

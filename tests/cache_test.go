package tests

import (
	"reflect"
	"testing"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Test caching functionality
func TestCachedParsing(t *testing.T) {
	parser := model.NewCachedParser[E2EUser](nil)

	userData := `{
		"id": 789,
		"username": "cacheduser",
		"email": "cached@example.com",
		"first_name": "Cached",
		"last_name": "User",
		"age": 32,
		"is_active": true,
		"profile": {
			"bio": "Testing caching",
			"location": "Cache City",
			"skills": ["Caching"],
			"languages": ["Binary"]
		},
		"created_at": "2023-07-01T08:00:00Z"
	}`

	// First parse (cache miss)
	user1, err := parser.Parse([]byte(userData))
	if err != nil {
		t.Fatalf("First parse failed: %v", err)
	}

	// Second parse (cache hit)
	user2, err := parser.Parse([]byte(userData))
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	// Results should be identical
	if !reflect.DeepEqual(user1, user2) {
		t.Error("Cached result should be identical to original")
	}

	if user1.Username != "cacheduser" {
		t.Errorf("Expected username 'cacheduser', got '%s'", user1.Username)
	}
}

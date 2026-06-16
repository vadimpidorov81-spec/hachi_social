package users

import (
	"testing"
	"time"
)

func TestUserUpdateProfileAndStatus(t *testing.T) {
	t.Parallel()

	id, err := ParseID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatal(err)
	}
	username, _ := NewUsername("alice")
	displayName, _ := NewDisplayName("Alice")
	timezone, _ := NewTimezone("Europe/Moscow")
	createdAt := time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)

	user, err := NewUser(id, username, displayName, timezone, createdAt)
	if err != nil {
		t.Fatal(err)
	}

	newDisplayName, _ := NewDisplayName("Alice Smith")
	bio, _ := NewBio("Go developer")
	newTimezone, _ := NewTimezone("UTC")
	updatedAt := createdAt.Add(time.Hour)
	user.UpdateProfile(newDisplayName, bio, newTimezone, updatedAt)
	user.SetStatus(StatusBlocked, updatedAt.Add(time.Hour))

	snapshot := user.Snapshot()
	if snapshot.DisplayName.String() != "Alice Smith" || snapshot.Bio.String() != "Go developer" {
		t.Fatalf("profile was not updated: %+v", snapshot)
	}
	if snapshot.Status != StatusBlocked {
		t.Fatalf("expected blocked status, got %s", snapshot.Status)
	}
}

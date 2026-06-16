package application

import (
	"context"
	"errors"
	"testing"
	"time"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
)

type fakeRepository struct {
	users map[domainusers.ID]*domainusers.User
}

func (r *fakeRepository) Create(_ context.Context, user *domainusers.User) error {
	for _, existing := range r.users {
		if existing.Snapshot().Username == user.Snapshot().Username {
			return ErrUsernameAlreadyTaken
		}
	}
	r.users[user.Snapshot().ID] = user
	return nil
}

func (r *fakeRepository) GetByID(_ context.Context, id domainusers.ID) (*domainusers.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *fakeRepository) GetByUsername(
	_ context.Context,
	username domainusers.Username,
) (*domainusers.User, error) {
	for _, user := range r.users {
		if user.Snapshot().Username == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *fakeRepository) UpdateProfile(_ context.Context, _ *domainusers.User) error {
	return nil
}

func (r *fakeRepository) UpdateStatusWithAudit(
	_ context.Context,
	_ domainusers.ID,
	_ *domainusers.User,
	_ domainusers.Status,
) error {
	return nil
}

type fixedIDGenerator struct {
	id domainusers.ID
}

func (g fixedIDGenerator) New() (domainusers.ID, error) {
	return g.id, nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestServiceCreateAndUpdateProfile(t *testing.T) {
	t.Parallel()

	id, _ := domainusers.ParseID("550e8400-e29b-41d4-a716-446655440000")
	repository := &fakeRepository{users: make(map[domainusers.ID]*domainusers.User)}
	service := NewService(
		repository,
		fixedIDGenerator{id: id},
		fixedClock{now: time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)},
	)

	created, err := service.Create(context.Background(), CreateUserCommand{
		Username:    "Alice",
		DisplayName: "Alice",
		Timezone:    "Europe/Moscow",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Username != "alice" {
		t.Fatalf("expected normalized username, got %q", created.Username)
	}

	bio := "Go developer"
	updated, err := service.UpdateProfile(context.Background(), id, UpdateProfileCommand{Bio: &bio})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Bio != bio {
		t.Fatalf("expected bio %q, got %q", bio, updated.Bio)
	}
}

func TestServiceSetStatusRequiresAdmin(t *testing.T) {
	t.Parallel()

	id, _ := domainusers.ParseID("550e8400-e29b-41d4-a716-446655440000")
	username, _ := domainusers.NewUsername("alice")
	displayName, _ := domainusers.NewDisplayName("Alice")
	timezone, _ := domainusers.NewTimezone("UTC")
	user, _ := domainusers.NewUser(id, username, displayName, timezone, time.Now())

	service := NewService(
		&fakeRepository{users: map[domainusers.ID]*domainusers.User{id: user}},
		fixedIDGenerator{id: id},
		fixedClock{now: time.Now()},
	)

	err := service.SetStatus(context.Background(), id, id, string(domainusers.StatusBlocked))
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

package application

import (
	"time"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
)

type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	Timezone    string    `json:"timezone"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PublicProfile struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
}

type CreateUserCommand struct {
	Username    string
	DisplayName string
	Timezone    string
}

type UpdateProfileCommand struct {
	DisplayName *string
	Bio         *string
	Timezone    *string
}

func toUser(user *domainusers.User) User {
	snapshot := user.Snapshot()
	return User{
		ID:          snapshot.ID.String(),
		Username:    snapshot.Username.String(),
		DisplayName: snapshot.DisplayName.String(),
		Bio:         snapshot.Bio.String(),
		Timezone:    snapshot.Timezone.String(),
		Role:        string(snapshot.Role),
		Status:      string(snapshot.Status),
		CreatedAt:   snapshot.CreatedAt,
		UpdatedAt:   snapshot.UpdatedAt,
	}
}

func toPublicProfile(user *domainusers.User) PublicProfile {
	snapshot := user.Snapshot()
	return PublicProfile{
		ID:          snapshot.ID.String(),
		Username:    snapshot.Username.String(),
		DisplayName: snapshot.DisplayName.String(),
		Bio:         snapshot.Bio.String(),
	}
}

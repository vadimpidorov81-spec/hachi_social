package users

import "time"

type User struct {
	id          ID
	username    Username
	displayName DisplayName
	bio         Bio
	timezone    Timezone
	role        Role
	status      Status
	createdAt   time.Time
	updatedAt   time.Time
}

type Snapshot struct {
	ID          ID
	Username    Username
	DisplayName DisplayName
	Bio         Bio
	Timezone    Timezone
	Role        Role
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewUser(
	id ID,
	username Username,
	displayName DisplayName,
	timezone Timezone,
	now time.Time,
) (*User, error) {
	if id.IsZero() {
		return nil, ErrInvalidID
	}

	bio, _ := NewBio("")
	now = now.UTC()
	return &User{
		id:          id,
		username:    username,
		displayName: displayName,
		bio:         bio,
		timezone:    timezone,
		role:        RoleUser,
		status:      StatusActive,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func Restore(snapshot Snapshot) (*User, error) {
	if snapshot.ID.IsZero() {
		return nil, ErrInvalidID
	}
	if _, err := ParseRole(string(snapshot.Role)); err != nil {
		return nil, err
	}
	if _, err := ParseStatus(string(snapshot.Status)); err != nil {
		return nil, err
	}
	return &User{
		id:          snapshot.ID,
		username:    snapshot.Username,
		displayName: snapshot.DisplayName,
		bio:         snapshot.Bio,
		timezone:    snapshot.Timezone,
		role:        snapshot.Role,
		status:      snapshot.Status,
		createdAt:   snapshot.CreatedAt.UTC(),
		updatedAt:   snapshot.UpdatedAt.UTC(),
	}, nil
}

func (u *User) UpdateProfile(displayName DisplayName, bio Bio, timezone Timezone, now time.Time) {
	u.displayName = displayName
	u.bio = bio
	u.timezone = timezone
	u.updatedAt = now.UTC()
}

func (u *User) SetStatus(status Status, now time.Time) {
	u.status = status
	u.updatedAt = now.UTC()
}

func (u *User) Snapshot() Snapshot {
	return Snapshot{
		ID:          u.id,
		Username:    u.username,
		DisplayName: u.displayName,
		Bio:         u.bio,
		Timezone:    u.timezone,
		Role:        u.role,
		Status:      u.status,
		CreatedAt:   u.createdAt,
		UpdatedAt:   u.updatedAt,
	}
}

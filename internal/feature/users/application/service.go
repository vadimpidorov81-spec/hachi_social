package application

import (
	"context"
	"errors"
	"fmt"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
)

type Service struct {
	repository  Repository
	idGenerator IDGenerator
	clock       Clock
}

func NewService(repository Repository, idGenerator IDGenerator, clock Clock) *Service {
	return &Service{
		repository:  repository,
		idGenerator: idGenerator,
		clock:       clock,
	}
}

func (s *Service) Create(ctx context.Context, command CreateUserCommand) (User, error) {
	username, err := domainusers.NewUsername(command.Username)
	if err != nil {
		return User{}, err
	}
	displayName, err := domainusers.NewDisplayName(command.DisplayName)
	if err != nil {
		return User{}, err
	}
	timezone, err := domainusers.NewTimezone(command.Timezone)
	if err != nil {
		return User{}, err
	}
	id, err := s.idGenerator.New()
	if err != nil {
		return User{}, fmt.Errorf("generate user id: %w", err)
	}
	user, err := domainusers.NewUser(id, username, displayName, timezone, s.clock.Now())
	if err != nil {
		return User{}, err
	}
	if err := s.repository.Create(ctx, user); err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return toUser(user), nil
}

func (s *Service) GetCurrent(ctx context.Context, id domainusers.ID) (User, error) {
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get current user: %w", err)
	}
	return toUser(user), nil
}

func (s *Service) GetPublicProfile(ctx context.Context, rawUsername string) (PublicProfile, error) {
	username, err := domainusers.NewUsername(rawUsername)
	if err != nil {
		return PublicProfile{}, ErrUserNotFound
	}
	user, err := s.repository.GetByUsername(ctx, username)
	if err != nil {
		return PublicProfile{}, fmt.Errorf("get public profile: %w", err)
	}
	if user.Snapshot().Status != domainusers.StatusActive {
		return PublicProfile{}, ErrUserNotFound
	}
	return toPublicProfile(user), nil
}

func (s *Service) UpdateProfile(
	ctx context.Context,
	id domainusers.ID,
	command UpdateProfileCommand,
) (User, error) {
	if command.DisplayName == nil && command.Bio == nil && command.Timezone == nil {
		return User{}, ErrEmptyUpdate
	}

	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user for profile update: %w", err)
	}
	current := user.Snapshot()

	displayName := current.DisplayName
	if command.DisplayName != nil {
		displayName, err = domainusers.NewDisplayName(*command.DisplayName)
		if err != nil {
			return User{}, err
		}
	}
	bio := current.Bio
	if command.Bio != nil {
		bio, err = domainusers.NewBio(*command.Bio)
		if err != nil {
			return User{}, err
		}
	}
	timezone := current.Timezone
	if command.Timezone != nil {
		timezone, err = domainusers.NewTimezone(*command.Timezone)
		if err != nil {
			return User{}, err
		}
	}

	user.UpdateProfile(displayName, bio, timezone, s.clock.Now())
	if err := s.repository.UpdateProfile(ctx, user); err != nil {
		return User{}, fmt.Errorf("update profile: %w", err)
	}
	return toUser(user), nil
}

func (s *Service) SetStatus(
	ctx context.Context,
	actorID domainusers.ID,
	targetID domainusers.ID,
	statusValue string,
) error {
	actor, err := s.repository.GetByID(ctx, actorID)
	if err != nil {
		return fmt.Errorf("get status actor: %w", err)
	}
	if actor.Snapshot().Role != domainusers.RoleAdmin {
		return ErrForbidden
	}
	if actorID == targetID && statusValue == string(domainusers.StatusBlocked) {
		return ErrCannotBlockSelf
	}

	status, err := domainusers.ParseStatus(statusValue)
	if err != nil {
		return err
	}
	target, err := s.repository.GetByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("get status target: %w", err)
	}
	previous := target.Snapshot().Status
	if previous == status {
		return nil
	}

	target.SetStatus(status, s.clock.Now())
	if err := s.repository.UpdateStatusWithAudit(ctx, actorID, target, previous); err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	return nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}

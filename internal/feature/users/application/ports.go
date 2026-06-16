package application

import (
	"context"
	"time"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
)

type Repository interface {
	Create(ctx context.Context, user *domainusers.User) error
	GetByID(ctx context.Context, id domainusers.ID) (*domainusers.User, error)
	GetByUsername(ctx context.Context, username domainusers.Username) (*domainusers.User, error)
	UpdateProfile(ctx context.Context, user *domainusers.User) error
	UpdateStatusWithAudit(
		ctx context.Context,
		actorID domainusers.ID,
		user *domainusers.User,
		previous domainusers.Status,
	) error
}

type IDGenerator interface {
	New() (domainusers.ID, error)
}

type Clock interface {
	Now() time.Time
}

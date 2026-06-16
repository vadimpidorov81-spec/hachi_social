package httptransport

import (
	"context"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
	"github.com/hachisocial/hachisocial/internal/feature/users/application"
)

type UseCase interface {
	GetCurrent(ctx context.Context, id domainusers.ID) (application.User, error)
	GetPublicProfile(ctx context.Context, username string) (application.PublicProfile, error)
	UpdateProfile(
		ctx context.Context,
		id domainusers.ID,
		command application.UpdateProfileCommand,
	) (application.User, error)
	SetStatus(
		ctx context.Context,
		actorID domainusers.ID,
		targetID domainusers.ID,
		status string,
	) error
}

type Principal struct {
	UserID domainusers.ID
	Role   domainusers.Role
}

type PrincipalProvider interface {
	Principal(ctx context.Context) (Principal, error)
}

type DenyPrincipalProvider struct{}

func (DenyPrincipalProvider) Principal(context.Context) (Principal, error) {
	return Principal{}, application.ErrUnauthorized
}

type StaticPrincipalProvider struct {
	principal Principal
}

func NewStaticPrincipalProvider(userID domainusers.ID, role domainusers.Role) StaticPrincipalProvider {
	return StaticPrincipalProvider{
		principal: Principal{
			UserID: userID,
			Role:   role,
		},
	}
}

func (p StaticPrincipalProvider) Principal(context.Context) (Principal, error) {
	return p.principal, nil
}

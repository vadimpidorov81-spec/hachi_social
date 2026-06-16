package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
	"github.com/hachisocial/hachisocial/internal/feature/users/application"
	generateddb "github.com/hachisocial/hachisocial/internal/generated/db"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *generateddb.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: generateddb.New(pool),
	}
}

func (r *Repository) Create(ctx context.Context, user *domainusers.User) error {
	snapshot := user.Snapshot()
	err := r.queries.CreateUser(ctx, generateddb.CreateUserParams{
		ID:          toDatabaseID(snapshot.ID),
		Username:    snapshot.Username.String(),
		DisplayName: snapshot.DisplayName.String(),
		Bio:         snapshot.Bio.String(),
		Timezone:    snapshot.Timezone.String(),
		Role:        string(snapshot.Role),
		Status:      string(snapshot.Status),
		CreatedAt:   pgtype.Timestamptz{Time: snapshot.CreatedAt, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: snapshot.UpdatedAt, Valid: true},
	})
	if isUniqueViolation(err) {
		return application.ErrUsernameAlreadyTaken
	}
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id domainusers.ID) (*domainusers.User, error) {
	user, err := r.queries.GetUserByID(ctx, toDatabaseID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user by id: %w", err)
	}
	return mapUser(user)
}

func (r *Repository) GetByUsername(
	ctx context.Context,
	username domainusers.Username,
) (*domainusers.User, error) {
	user, err := r.queries.GetUserByUsername(ctx, username.String())
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user by username: %w", err)
	}
	return mapUser(user)
}

func (r *Repository) UpdateProfile(ctx context.Context, user *domainusers.User) error {
	snapshot := user.Snapshot()
	rowsAffected, err := r.queries.UpdateUserProfile(ctx, generateddb.UpdateUserProfileParams{
		ID:          toDatabaseID(snapshot.ID),
		DisplayName: snapshot.DisplayName.String(),
		Bio:         snapshot.Bio.String(),
		Timezone:    snapshot.Timezone.String(),
		UpdatedAt:   pgtype.Timestamptz{Time: snapshot.UpdatedAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	if rowsAffected == 0 {
		return application.ErrUserNotFound
	}
	return nil
}

func (r *Repository) UpdateStatusWithAudit(
	ctx context.Context,
	actorID domainusers.ID,
	user *domainusers.User,
	previous domainusers.Status,
) error {
	transaction, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin status transaction: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	snapshot := user.Snapshot()
	command, err := transaction.Exec(ctx, `
		UPDATE users
		SET status = $2, updated_at = $3
		WHERE id = $1 AND status = $4
	`, snapshot.ID.String(), snapshot.Status, snapshot.UpdatedAt, previous)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	if command.RowsAffected() == 0 {
		return application.ErrUserNotFound
	}

	_, err = transaction.Exec(ctx, `
		INSERT INTO audit_log (
			id, actor_id, target_id, action, old_value, new_value, created_at
		) VALUES (gen_random_uuid(), $1, $2, 'user.status_changed', $3, $4, $5)
	`,
		actorID.String(),
		snapshot.ID.String(),
		previous,
		snapshot.Status,
		snapshot.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert status audit: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit status transaction: %w", err)
	}
	return nil
}

func mapUser(databaseUser generateddb.User) (*domainusers.User, error) {
	if !databaseUser.ID.Valid || !databaseUser.CreatedAt.Valid || !databaseUser.UpdatedAt.Valid {
		return nil, errors.New("map user: required database value is null")
	}

	id := domainusers.IDFromBytes(databaseUser.ID.Bytes)
	username, err := domainusers.NewUsername(databaseUser.Username)
	if err != nil {
		return nil, fmt.Errorf("map username: %w", err)
	}
	displayName, err := domainusers.NewDisplayName(databaseUser.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("map display name: %w", err)
	}
	bio, err := domainusers.NewBio(databaseUser.Bio)
	if err != nil {
		return nil, fmt.Errorf("map bio: %w", err)
	}
	timezone, err := domainusers.NewTimezone(databaseUser.Timezone)
	if err != nil {
		return nil, fmt.Errorf("map timezone: %w", err)
	}
	role, err := domainusers.ParseRole(databaseUser.Role)
	if err != nil {
		return nil, fmt.Errorf("map role: %w", err)
	}
	status, err := domainusers.ParseStatus(databaseUser.Status)
	if err != nil {
		return nil, fmt.Errorf("map status: %w", err)
	}

	user, err := domainusers.Restore(domainusers.Snapshot{
		ID:          id,
		Username:    username,
		DisplayName: displayName,
		Bio:         bio,
		Timezone:    timezone,
		Role:        role,
		Status:      status,
		CreatedAt:   databaseUser.CreatedAt.Time,
		UpdatedAt:   databaseUser.UpdatedAt.Time,
	})
	if err != nil {
		return nil, fmt.Errorf("restore user: %w", err)
	}
	return user, nil
}

func toDatabaseID(id domainusers.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: id.Bytes(), Valid: true}
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

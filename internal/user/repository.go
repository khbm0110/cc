package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrPlanNotFound = errors.New("plan not found")
)

// Repository defines the interface for user and plan persistence.
type Repository interface {
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetUserWithPlan(ctx context.Context, userID int64) (*UserWithPlan, error)
	GetPlanByID(ctx context.Context, id int64) (*Plan, error)
	ListActiveUsers(ctx context.Context) ([]User, error)
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, name, plan_id, api_key_encrypted, secret_key_encrypted, created_at, updated_at
		FROM users WHERE id = $1`

	var u User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.PlanID, &u.APIKeyEncrypted, &u.SecretKeyEncrypted,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) GetUserWithPlan(ctx context.Context, userID int64) (*UserWithPlan, error) {
	query := `
		SELECT u.id, u.name, u.plan_id, u.api_key_encrypted, u.secret_key_encrypted,
		       u.created_at, u.updated_at,
		       p.id, p.name, p.max_exposure_ratio, p.order_limit_per_min,
		       p.created_at, p.updated_at
		FROM users u
		JOIN plans p ON u.plan_id = p.id
		WHERE u.id = $1`

	var uwp UserWithPlan
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&uwp.User.ID, &uwp.User.Name, &uwp.User.PlanID,
		&uwp.User.APIKeyEncrypted, &uwp.User.SecretKeyEncrypted,
		&uwp.User.CreatedAt, &uwp.User.UpdatedAt,
		&uwp.Plan.ID, &uwp.Plan.Name, &uwp.Plan.MaxExposureRatio,
		&uwp.Plan.OrderLimitPerMin,
		&uwp.Plan.CreatedAt, &uwp.Plan.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user with plan: %w", err)
	}
	return &uwp, nil
}

func (r *PostgresRepository) GetPlanByID(ctx context.Context, id int64) (*Plan, error) {
	query := `
		SELECT id, name, max_exposure_ratio, order_limit_per_min, created_at, updated_at
		FROM plans WHERE id = $1`

	var p Plan
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.MaxExposureRatio, &p.OrderLimitPerMin,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("get plan by id: %w", err)
	}
	return &p, nil
}

func (r *PostgresRepository) ListActiveUsers(ctx context.Context) ([]User, error) {
	query := `
		SELECT id, name, plan_id, api_key_encrypted, secret_key_encrypted, created_at, updated_at
		FROM users ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.PlanID, &u.APIKeyEncrypted, &u.SecretKeyEncrypted,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

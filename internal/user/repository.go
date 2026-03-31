package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrPlanNotFound = errors.New("plan not found")
)

// CreateUserParams contains fields needed to create a new user.
type CreateUserParams struct {
	Email        string
	PasswordHash string
	Name         string
	Role         string
}

// Repository defines the interface for user and plan persistence.
type Repository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetUserWithPlan(ctx context.Context, userID int64) (*UserWithPlan, error)
	GetPlanByID(ctx context.Context, id int64) (*Plan, error)
	ListActiveUsers(ctx context.Context) ([]User, error)
	CreateUser(ctx context.Context, params CreateUserParams) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByRefreshToken(ctx context.Context, token string) (*User, error)
	UpdateRefreshToken(ctx context.Context, userID int64, token string, expiry time.Time) error
	UpdateAPIKeys(ctx context.Context, userID int64, apiKey, secretKey string) error
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, role, plan_id, api_key_encrypted, secret_key_encrypted, refresh_token, refresh_token_expiry, created_at, updated_at
		FROM users WHERE id = $1`

	var u User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.PlanID, &u.APIKeyEncrypted, &u.SecretKeyEncrypted,
		&u.RefreshToken, &u.RefreshTokenExpiry, &u.CreatedAt, &u.UpdatedAt,
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
		SELECT u.id, u.email, u.password_hash, u.name, u.role, u.plan_id, u.api_key_encrypted, u.secret_key_encrypted,
		       u.refresh_token, u.refresh_token_expiry, u.created_at, u.updated_at,
		       p.id, p.name, p.max_exposure_ratio, p.order_limit_per_min,
		       p.created_at, p.updated_at
		FROM users u
		LEFT JOIN plans p ON u.plan_id = p.id
		WHERE u.id = $1`

	var uwp UserWithPlan
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&uwp.User.ID, &uwp.User.Email, &uwp.User.PasswordHash, &uwp.User.Name, &uwp.User.Role, &uwp.User.PlanID,
		&uwp.User.APIKeyEncrypted, &uwp.User.SecretKeyEncrypted, &uwp.User.RefreshToken, &uwp.User.RefreshTokenExpiry,
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
		SELECT id, email, password_hash, name, role, plan_id, api_key_encrypted, secret_key_encrypted, refresh_token, refresh_token_expiry, created_at, updated_at
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
			&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.PlanID, &u.APIKeyEncrypted, &u.SecretKeyEncrypted,
			&u.RefreshToken, &u.RefreshTokenExpiry, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *PostgresRepository) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, email, password_hash, name, role, plan_id, api_key_encrypted, secret_key_encrypted, refresh_token, refresh_token_expiry, created_at, updated_at`

	var u User
	err := r.db.QueryRowContext(ctx, query, params.Email, params.PasswordHash, params.Name, params.Role).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.PlanID, &u.APIKeyEncrypted, &u.SecretKeyEncrypted,
		&u.RefreshToken, &u.RefreshTokenExpiry, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, role, plan_id, api_key_encrypted, secret_key_encrypted, refresh_token, refresh_token_expiry, created_at, updated_at
		FROM users WHERE email = $1`

	var u User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.PlanID, &u.APIKeyEncrypted, &u.SecretKeyEncrypted,
		&u.RefreshToken, &u.RefreshTokenExpiry, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) GetByRefreshToken(ctx context.Context, token string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, role, plan_id, api_key_encrypted, secret_key_encrypted, refresh_token, refresh_token_expiry, created_at, updated_at
		FROM users WHERE refresh_token = $1`

	var u User
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.PlanID, &u.APIKeyEncrypted, &u.SecretKeyEncrypted,
		&u.RefreshToken, &u.RefreshTokenExpiry, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by refresh token: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) UpdateRefreshToken(ctx context.Context, userID int64, token string, expiry time.Time) error {
	query := `UPDATE users SET refresh_token = $1, refresh_token_expiry = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, token, expiry, userID)
	if err != nil {
		return fmt.Errorf("update refresh token: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateAPIKeys(ctx context.Context, userID int64, apiKey, secretKey string) error {
	query := `UPDATE users SET api_key_encrypted = $1, secret_key_encrypted = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, apiKey, secretKey, userID)
	if err != nil {
		return fmt.Errorf("update api keys: %w", err)
	}
	return nil
}
package user

import "time"

// Plan represents a subscription plan with trading limits.
type Plan struct {
	ID                int64   `db:"id" json:"id"`
	Name              string  `db:"name" json:"name"`
	MaxExposureRatio  float64 `db:"max_exposure_ratio" json:"max_exposure_ratio"`
	OrderLimitPerMin  int     `db:"order_limit_per_min" json:"order_limit_per_min"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// User represents a platform user with encrypted API credentials.
type User struct {
	ID                 int64  `db:"id" json:"id"`
	Name               string `db:"name" json:"name"`
	PlanID             int64  `db:"plan_id" json:"plan_id"`
	APIKeyEncrypted    []byte `db:"api_key_encrypted" json:"-"`
	SecretKeyEncrypted []byte `db:"secret_key_encrypted" json:"-"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

// UserWithPlan combines user data with their associated plan.
type UserWithPlan struct {
	User User
	Plan Plan
}

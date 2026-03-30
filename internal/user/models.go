package user

import "time"

// Plan represents a subscription plan with trading limits and virtual balance.
type Plan struct {
	ID                int64     `db:"id" json:"id"`
	Name              string    `db:"name" json:"name"`
	MaxExposureRatio  float64   `db:"max_exposure_ratio" json:"max_exposure_ratio"`
	OrderLimitPerMin  int       `db:"order_limit_per_min" json:"order_limit_per_min"`
	TraderID          *int64    `db:"trader_id" json:"trader_id,omitempty"`
	VirtualBalance    float64   `db:"virtual_balance" json:"virtual_balance"`
	SubscriptionPrice float64   `db:"subscription_price" json:"subscription_price"`
	MinInvestment     float64   `db:"min_investment" json:"min_investment"`
	Description       string    `db:"description" json:"description"`
	IsActive          bool      `db:"is_active" json:"is_active"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// Trader represents a professional trader assigned by admin.
type Trader struct {
	ID              int64     `db:"id" json:"id"`
	UserID          *int64    `db:"user_id" json:"user_id,omitempty"`
	Name            string    `db:"name" json:"name"`
	Email           string    `db:"email" json:"email"`
	AvatarURL       string    `db:"avatar_url" json:"avatar_url"`
	IsActive        bool      `db:"is_active" json:"is_active"`
	MaxTradesPerDay int       `db:"max_trades_per_day" json:"max_trades_per_day"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// UserSubscription represents a user's subscription to a plan.
type UserSubscription struct {
	ID             int64      `db:"id" json:"id"`
	UserID         int64      `db:"user_id" json:"user_id"`
	PlanID         int64      `db:"plan_id" json:"plan_id"`
	SubscribedAt   time.Time  `db:"subscribed_at" json:"subscribed_at"`
	ExpiresAt      *time.Time `db:"expires_at" json:"expires_at"`
	IsActive       bool       `db:"is_active" json:"is_active"`
	MonthlyFeePaid float64    `db:"monthly_fee_paid" json:"monthly_fee_paid"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// User represents a platform user with encrypted API credentials.
type User struct {
	ID                 int64      `json:"id" db:"id"`
	Email              string     `json:"email" db:"email"`
	PasswordHash       string     `json:"-" db:"password_hash"`
	Name               string     `json:"name" db:"name"`
	Role               string     `json:"role" db:"role"`
	PlanID             *int64     `json:"plan_id,omitempty" db:"plan_id"`
	APIKeyEncrypted    []byte     `json:"-" db:"api_key_encrypted"`
	SecretKeyEncrypted []byte     `json:"-" db:"secret_key_encrypted"`
	RefreshToken       string     `json:"-" db:"refresh_token"`
	RefreshTokenExpiry time.Time  `json:"-" db:"refresh_token_expiry"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// UserWithPlan combines user data with their associated plan.
type UserWithPlan struct {
	User         User
	Plan         Plan
	Subscription *UserSubscription
}

// UserWithFullDetails combines user with plan, subscription, and trader info.
type UserWithFullDetails struct {
	User         User
	Plan         Plan
	Subscription *UserSubscription
	Trader       *Trader
}
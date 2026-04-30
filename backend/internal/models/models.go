package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Balance      float64   `json:"balance" db:"balance"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Sport struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Slug     string `json:"slug" db:"slug"`
	Icon     string `json:"icon" db:"icon"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type Event struct {
	ID             uuid.UUID `json:"id" db:"id"`
	SportID        int       `json:"sport_id" db:"sport_id"`
	HomeTeam       string    `json:"home_team" db:"home_team"`
	AwayTeam       string    `json:"away_team" db:"away_team"`
	StartTime      time.Time `json:"start_time" db:"start_time"`
	Status         string    `json:"status" db:"status"`
	FinalScoreHome *int      `json:"final_score_home,omitempty" db:"final_score_home"`
	FinalScoreAway *int      `json:"final_score_away,omitempty" db:"final_score_away"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	// Joined fields
	SportName    string `json:"sport_name,omitempty"`
	SportSlug    string `json:"sport_slug,omitempty"`
	SportIcon    string `json:"sport_icon,omitempty"`
	MarketsCount int    `json:"markets_count,omitempty"`
}

type Market struct {
	ID               uuid.UUID `json:"id" db:"id"`
	EventID          uuid.UUID `json:"event_id" db:"event_id"`
	MarketType       string    `json:"market_type" db:"market_type"`
	Name             string    `json:"name" db:"name"`
	Line             *float64  `json:"line,omitempty" db:"line"`
	Status           string    `json:"status" db:"status"`
	MarginPercentage float64   `json:"margin_percentage" db:"margin_percentage"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	// Joined fields
	Odds []Odd `json:"odds,omitempty"`
}

type Odd struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	MarketID         uuid.UUID  `json:"market_id" db:"market_id"`
	Outcome          string     `json:"outcome" db:"outcome"`
	InitialOdds      float64    `json:"initial_odds" db:"initial_odds"`
	CurrentOdds      float64    `json:"current_odds" db:"current_odds"`
	TotalStake       float64    `json:"total_stake" db:"total_stake"`
	BetCount         int        `json:"bet_count" db:"bet_count"`
	LastCalculatedAt *time.Time `json:"last_calculated_at,omitempty" db:"last_calculated_at"`
	IsManualOverride bool       `json:"is_manual_override" db:"is_manual_override"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type MarketPool struct {
	ID           uuid.UUID `json:"id" db:"id"`
	MarketID     uuid.UUID `json:"market_id" db:"market_id"`
	TotalPool    float64   `json:"total_pool" db:"total_pool"`
	HouseMargin  float64   `json:"house_margin" db:"house_margin"`
	Liability    float64   `json:"liability" db:"liability"`
	CalculatedAt time.Time `json:"calculated_at" db:"calculated_at"`
}

type Bet struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	BetType         string     `json:"bet_type" db:"bet_type"`
	Stake           float64    `json:"stake" db:"stake"`
	PotentialReturn float64    `json:"potential_return" db:"potential_return"`
	ActualReturn    float64    `json:"actual_return" db:"actual_return"`
	Status          string     `json:"status" db:"status"`
	PlacedAt        time.Time  `json:"placed_at" db:"placed_at"`
	SettledAt       *time.Time `json:"settled_at,omitempty" db:"settled_at"`
	// Joined fields
	Legs []BetLeg `json:"legs,omitempty"`
}

type BetLeg struct {
	ID             uuid.UUID `json:"id" db:"id"`
	BetID          uuid.UUID `json:"bet_id" db:"bet_id"`
	MarketID       uuid.UUID `json:"market_id" db:"market_id"`
	OddID          uuid.UUID `json:"odd_id" db:"odd_id"`
	Outcome        string    `json:"outcome" db:"outcome"`
	LockedOddValue float64   `json:"locked_odd_value" db:"locked_odd_value"`
	Result         string    `json:"result" db:"result"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	// Joined fields
	EventName  string `json:"event_name,omitempty"`
	MarketType string `json:"market_type,omitempty"`
}

type Transaction struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Type          string     `json:"type" db:"type"`
	Amount        float64    `json:"amount" db:"amount"`
	BalanceBefore float64    `json:"balance_before" db:"balance_before"`
	BalanceAfter  float64    `json:"balance_after" db:"balance_after"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty" db:"reference_id"`
	Status        string     `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type OddsHistory struct {
	ID         uuid.UUID `json:"id" db:"id"`
	OddID      uuid.UUID `json:"odd_id" db:"odd_id"`
	OddsValue  float64   `json:"odds_value" db:"odds_value"`
	TotalStake float64   `json:"total_stake" db:"total_stake"`
	BetCount   int       `json:"bet_count" db:"bet_count"`
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
}

type AdminLog struct {
	ID         uuid.UUID   `json:"id" db:"id"`
	AdminID    uuid.UUID   `json:"admin_id" db:"admin_id"`
	Action     string      `json:"action" db:"action"`
	EntityType *string     `json:"entity_type,omitempty" db:"entity_type"`
	EntityID   *uuid.UUID  `json:"entity_id,omitempty" db:"entity_id"`
	Details    interface{} `json:"details,omitempty" db:"details"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}

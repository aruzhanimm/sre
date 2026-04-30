-- BetKZ Database Schema
-- Complete schema with dynamic odds system

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Users table
-- ============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    balance DECIMAL(15,2) DEFAULT 0.00,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- Sports categories
-- ============================================
CREATE TABLE sports (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    icon VARCHAR(50) DEFAULT '',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- Events (matches/games)
-- ============================================
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sport_id INTEGER NOT NULL REFERENCES sports(id) ON DELETE RESTRICT,
    home_team VARCHAR(255) NOT NULL,
    away_team VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) DEFAULT 'upcoming' CHECK (status IN ('upcoming', 'live', 'finished', 'cancelled', 'settled')),
    final_score_home INTEGER,
    final_score_away INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_events_sport_id ON events(sport_id);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_time ON events(start_time);

-- ============================================
-- Markets (betting markets for events)
-- ============================================
CREATE TABLE markets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    market_type VARCHAR(50) NOT NULL CHECK (market_type IN ('1x2', 'over_under', 'both_teams_score', 'double_chance', 'handicap', 'custom')),
    name VARCHAR(255) NOT NULL DEFAULT '',
    line DECIMAL(5,2),  -- For over/under, handicap etc.
    status VARCHAR(50) DEFAULT 'open' CHECK (status IN ('open', 'locked', 'settled', 'cancelled')),
    margin_percentage DECIMAL(5,2) DEFAULT 5.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_markets_event_id ON markets(event_id);
CREATE INDEX idx_markets_status ON markets(status);

-- ============================================
-- Odds with dynamic calculation support
-- ============================================
CREATE TABLE odds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    outcome VARCHAR(100) NOT NULL,
    initial_odds DECIMAL(10,2) NOT NULL CHECK (initial_odds >= 1.01),
    current_odds DECIMAL(10,2) NOT NULL CHECK (current_odds >= 1.01),
    total_stake DECIMAL(15,2) DEFAULT 0.00,
    bet_count INTEGER DEFAULT 0,
    last_calculated_at TIMESTAMP WITH TIME ZONE,
    is_manual_override BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(market_id, outcome)
);

CREATE INDEX idx_odds_market_id ON odds(market_id);
CREATE INDEX idx_odds_current_odds ON odds(current_odds);
CREATE INDEX idx_odds_total_stake ON odds(total_stake);

-- ============================================
-- Market pools tracking
-- ============================================
CREATE TABLE market_pools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    market_id UUID UNIQUE NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    total_pool DECIMAL(15,2) DEFAULT 0.00,
    house_margin DECIMAL(5,2) DEFAULT 5.00,
    liability DECIMAL(15,2) DEFAULT 0.00,
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- Bets placed by users
-- ============================================
CREATE TABLE bets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    bet_type VARCHAR(20) NOT NULL CHECK (bet_type IN ('single', 'accumulator')),
    stake DECIMAL(15,2) NOT NULL CHECK (stake >= 0.50),
    potential_return DECIMAL(15,2) NOT NULL,
    actual_return DECIMAL(15,2) DEFAULT 0.00,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'won', 'lost', 'cancelled')),
    placed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    settled_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_bets_user_id ON bets(user_id);
CREATE INDEX idx_bets_status ON bets(status);
CREATE INDEX idx_bets_placed_at ON bets(placed_at);

-- ============================================
-- Bet legs (individual selections in a bet)
-- ============================================
CREATE TABLE bet_legs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bet_id UUID NOT NULL REFERENCES bets(id) ON DELETE CASCADE,
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE RESTRICT,
    odd_id UUID NOT NULL REFERENCES odds(id) ON DELETE RESTRICT,
    outcome VARCHAR(100) NOT NULL,
    locked_odd_value DECIMAL(10,2) NOT NULL,
    result VARCHAR(50) DEFAULT 'pending' CHECK (result IN ('pending', 'won', 'lost', 'push')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_bet_legs_bet_id ON bet_legs(bet_id);
CREATE INDEX idx_bet_legs_market_id ON bet_legs(market_id);

-- ============================================
-- Transaction history
-- ============================================
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('deposit', 'withdraw', 'bet_placed', 'bet_won', 'bet_lost', 'bet_cancelled')),
    amount DECIMAL(15,2) NOT NULL,
    balance_before DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    reference_id UUID,
    status VARCHAR(50) DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);

-- ============================================
-- Odds history for analytics
-- ============================================
CREATE TABLE odds_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    odd_id UUID NOT NULL REFERENCES odds(id) ON DELETE CASCADE,
    odds_value DECIMAL(10,2) NOT NULL,
    total_stake DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    bet_count INTEGER NOT NULL DEFAULT 0,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_odds_history_odd_id ON odds_history(odd_id);
CREATE INDEX idx_odds_history_recorded_at ON odds_history(recorded_at);

-- ============================================
-- Admin activity log
-- ============================================
CREATE TABLE admin_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_admin_logs_admin_id ON admin_logs(admin_id);
CREATE INDEX idx_admin_logs_created_at ON admin_logs(created_at);

-- ============================================
-- Triggers
-- ============================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_events_updated_at BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_markets_updated_at BEFORE UPDATE ON markets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_odds_updated_at BEFORE UPDATE ON odds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Auto-log odds changes to odds_history
CREATE OR REPLACE FUNCTION log_odds_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.current_odds IS DISTINCT FROM NEW.current_odds THEN
        INSERT INTO odds_history (odd_id, odds_value, total_stake, bet_count)
        VALUES (NEW.id, NEW.current_odds, NEW.total_stake, NEW.bet_count);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_odds_history AFTER UPDATE ON odds
    FOR EACH ROW EXECUTE FUNCTION log_odds_change();

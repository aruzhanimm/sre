package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"BetKZ/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ========================================================================
// Test Suite
// ========================================================================

type E2ESuite struct {
	suite.Suite
	db     *pgxpool.Pool
	rdb    *redis.Client
	router *gin.Engine
	ts     *httptest.Server

	// Tokens stored during tests
	userToken    string
	refreshToken string
	adminToken   string
	userID       string
	adminID      string
}

func TestE2E(t *testing.T) {
	if os.Getenv("E2E_TESTS") == "" {
		t.Skip("Skipping E2E tests. Set E2E_TESTS=1 to run.")
	}
	suite.Run(t, new(E2ESuite))
}

func (s *E2ESuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://betkz:betkz_dev_pass@localhost:5432/betkz_test?sslmode=disable"
	}

	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/1"
	}

	// Connect
	s.db = database.Connect(dbURL)
	s.rdb = database.ConnectRedis(redisURL)

	// Run migrations
	require.NoError(s.T(), runMigrations(s.db))

	// Setup router
	s.router = SetupRouter(s.db, s.rdb, "test-secret-key", "*")
	s.ts = httptest.NewServer(s.router)
}

func (s *E2ESuite) TearDownSuite() {
	s.cleanDB()
	s.ts.Close()
	s.db.Close()
	s.rdb.Close()
}

func (s *E2ESuite) cleanDB() {
	tables := []string{
		"odds_history", "admin_logs", "bet_legs", "transactions", "bets",
		"market_pools", "odds", "markets", "events", "users",
	}
	for _, t := range tables {
		s.db.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s", t))
	}
}

// ========================================================================
// Helpers
// ========================================================================

type apiResponse map[string]interface{}

func (s *E2ESuite) request(method, path string, body interface{}, token string) (*http.Response, apiResponse) {
	var bodyReader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, s.ts.URL+path, bodyReader)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)

	var result apiResponse
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	_ = json.Unmarshal(data, &result)

	return resp, result
}

func (s *E2ESuite) requestArray(method, path string, token string) (*http.Response, []byte) {
	req, _ := http.NewRequest(method, s.ts.URL+path, nil)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, _ := http.DefaultClient.Do(req)
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, data
}

// uniqueEmail generates a unique email for each test run
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@test.com", prefix, time.Now().UnixNano())
}

// ========================================================================
// 1. Health Tests
// ========================================================================

func (s *E2ESuite) Test01_Health() {
	t := s.T()

	resp, body := s.request("GET", "/health", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", body["status"])

	resp, body = s.request("GET", "/health/db", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", body["status"])

	resp, body = s.request("GET", "/health/redis", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", body["status"])
}

// ========================================================================
// 2. Auth Flow Tests
// ========================================================================

func (s *E2ESuite) Test02_Auth_Register() {
	t := s.T()

	// Register user
	resp, body := s.request("POST", "/api/auth/register", map[string]string{
		"email": "testuser@betkz.test", "password": "Test1234!",
	}, "")
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.NotEmpty(t, body["access_token"])
	assert.NotEmpty(t, body["refresh_token"])

	user := body["user"].(map[string]interface{})
	assert.Equal(t, "testuser@betkz.test", user["email"])
	assert.Equal(t, "user", user["role"])
	s.userID = user["id"].(string)
	s.userToken = body["access_token"].(string)
	s.refreshToken = body["refresh_token"].(string)
}

func (s *E2ESuite) Test03_Auth_Register_Duplicate() {
	t := s.T()

	resp, body := s.request("POST", "/api/auth/register", map[string]string{
		"email": "testuser@betkz.test", "password": "Test1234!",
	}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, body["error"].(string), "exists")
}

func (s *E2ESuite) Test04_Auth_Login() {
	t := s.T()

	resp, body := s.request("POST", "/api/auth/login", map[string]string{
		"email": "testuser@betkz.test", "password": "Test1234!",
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, body["access_token"])
	s.userToken = body["access_token"].(string)
	s.refreshToken = body["refresh_token"].(string)
}

func (s *E2ESuite) Test05_Auth_Login_WrongPassword() {
	t := s.T()

	resp, _ := s.request("POST", "/api/auth/login", map[string]string{
		"email": "testuser@betkz.test", "password": "WrongPass!",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func (s *E2ESuite) Test06_Auth_RefreshToken() {
	t := s.T()

	resp, body := s.request("POST", "/api/auth/refresh", map[string]string{
		"refresh_token": s.refreshToken,
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, body["access_token"])
	s.userToken = body["access_token"].(string)
}

func (s *E2ESuite) Test07_Auth_Profile() {
	t := s.T()

	resp, body := s.request("GET", "/api/auth/profile", nil, s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	user := body["user"].(map[string]interface{})
	assert.Equal(t, "testuser@betkz.test", user["email"])
}

func (s *E2ESuite) Test08_Auth_Profile_Unauthorized() {
	t := s.T()

	resp, _ := s.request("GET", "/api/auth/profile", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ========================================================================
// 3. Admin Setup — make testuser admin
// ========================================================================

func (s *E2ESuite) Test09_Admin_Setup() {
	t := s.T()

	// Promote user to admin directly in DB
	_, err := s.db.Exec(context.Background(),
		"UPDATE users SET role = 'admin' WHERE email = 'testuser@betkz.test'")
	require.NoError(t, err)

	// Re-login to get admin token
	resp, body := s.request("POST", "/api/auth/login", map[string]string{
		"email": "testuser@betkz.test", "password": "Test1234!",
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	s.adminToken = body["access_token"].(string)
	s.adminID = s.userID

	// Register a regular user for betting tests
	resp, body = s.request("POST", "/api/auth/register", map[string]string{
		"email": "bettor@betkz.test", "password": "Bettor1234!",
	}, "")
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	user := body["user"].(map[string]interface{})
	s.userID = user["id"].(string)
	s.userToken = body["access_token"].(string)
}

// ========================================================================
// 4. Sports & Events Tests
// ========================================================================

func (s *E2ESuite) Test10_Sports_ListEmpty() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/sports", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Seed data may or may not exist
	_ = data
}

func (s *E2ESuite) Test11_Admin_CreateSeedSport() {
	// Insert a test sport directly
	_, err := s.db.Exec(context.Background(),
		`INSERT INTO sports (name, slug, icon, is_active) VALUES ('Football', 'football', '⚽', true) ON CONFLICT DO NOTHING`)
	require.NoError(s.T(), err)
}

func (s *E2ESuite) Test12_Admin_CreateEvent() {
	t := s.T()

	startTime := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	resp, body := s.request("POST", "/api/admin/events", map[string]interface{}{
		"sport_id":   1,
		"home_team":  "Team Alpha",
		"away_team":  "Team Beta",
		"start_time": startTime,
	}, s.adminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.NotEmpty(t, body["id"])
}

func (s *E2ESuite) Test13_Admin_CreateEvent_Unauthorized() {
	t := s.T()

	resp, _ := s.request("POST", "/api/admin/events", map[string]interface{}{
		"sport_id": 1, "home_team": "X", "away_team": "Y",
		"start_time": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	}, s.userToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func (s *E2ESuite) Test14_Events_List() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/events", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	events := result["events"].([]interface{})
	assert.GreaterOrEqual(t, len(events), 1)
}

func (s *E2ESuite) Test15_Events_ListByStatus() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/events?status=upcoming", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	events := result["events"].([]interface{})
	for _, e := range events {
		ev := e.(map[string]interface{})
		assert.Equal(t, "upcoming", ev["status"])
	}
}

func (s *E2ESuite) Test16_Events_GetByID() {
	t := s.T()

	// Get first event
	_, data := s.requestArray("GET", "/api/events?limit=1", "")
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	events := result["events"].([]interface{})
	require.NotEmpty(t, events)

	eventID := events[0].(map[string]interface{})["id"].(string)

	resp, body := s.request("GET", "/api/events/"+eventID, nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, eventID, body["id"])
	assert.Equal(t, "Team Alpha", body["home_team"])
}

// ========================================================================
// 5. Markets & Odds Tests
// ========================================================================

var testEventID string
var testMarketID string
var testOddIDs []string

func (s *E2ESuite) Test17_Admin_CreateMarket() {
	t := s.T()

	// Get event ID
	_, data := s.requestArray("GET", "/api/events?limit=1", "")
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	events := result["events"].([]interface{})
	require.NotEmpty(t, events)
	testEventID = events[0].(map[string]interface{})["id"].(string)

	resp, body := s.request("POST", "/api/admin/markets", map[string]interface{}{
		"event_id":    testEventID,
		"market_type": "1x2",
		"name":        "Match Result",
		"outcomes": []map[string]interface{}{
			{"outcome": "Home", "initial_odds": 2.10},
			{"outcome": "Draw", "initial_odds": 3.40},
			{"outcome": "Away", "initial_odds": 3.20},
		},
	}, s.adminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.NotEmpty(t, body["id"])
	testMarketID = body["id"].(string)
}

func (s *E2ESuite) Test18_EventMarkets() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/events/"+testEventID+"/markets", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	markets := result["markets"].([]interface{})
	assert.GreaterOrEqual(t, len(markets), 1)

	market := markets[0].(map[string]interface{})
	assert.Equal(t, "Match Result", market["name"])

	odds := market["odds"].([]interface{})
	assert.Equal(t, 3, len(odds))

	// Store odd IDs for betting
	testOddIDs = nil
	for _, o := range odds {
		odd := o.(map[string]interface{})
		testOddIDs = append(testOddIDs, odd["id"].(string))
	}
}

func (s *E2ESuite) Test19_MarketOdds() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/markets/"+testMarketID+"/odds", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	odds := result["odds"].([]interface{})
	assert.Equal(t, 3, len(odds))
}

// ========================================================================
// 6. Admin Deposit (fund the test bettor)
// ========================================================================

func (s *E2ESuite) Test20_Admin_Deposit() {
	t := s.T()

	resp, body := s.request("POST", "/api/admin/users/"+s.userID+"/deposit", map[string]interface{}{
		"amount": 1000.00,
	}, s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "deposit successful", body["message"])
}

func (s *E2ESuite) Test21_User_BalanceAfterDeposit() {
	t := s.T()

	resp, body := s.request("GET", "/api/auth/profile", nil, s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	user := body["user"].(map[string]interface{})
	assert.Equal(t, 1000.0, user["balance"])
}

// ========================================================================
// 7. Betting Flow Tests
// ========================================================================

var testBetID string

func (s *E2ESuite) Test22_PlaceBet() {
	t := s.T()
	require.NotEmpty(t, testOddIDs, "No odds available for betting")

	// Get the first odd's details
	_, data := s.requestArray("GET", "/api/events/"+testEventID+"/markets", "")
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	markets := result["markets"].([]interface{})
	market := markets[0].(map[string]interface{})
	odds := market["odds"].([]interface{})
	firstOdd := odds[0].(map[string]interface{})

	resp, body := s.request("POST", "/api/bets", map[string]interface{}{
		"stake": 50.0,
		"selections": []map[string]interface{}{
			{
				"market_id": testMarketID,
				"odd_id":    firstOdd["id"],
				"outcome":   firstOdd["outcome"],
			},
		},
	}, s.userToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.NotEmpty(t, body["id"])
	assert.Equal(t, "pending", body["status"])
	assert.Equal(t, 50.0, body["stake"])
	testBetID = body["id"].(string)
}

func (s *E2ESuite) Test23_PlaceBet_InsufficientBalance() {
	t := s.T()

	resp, body := s.request("POST", "/api/bets", map[string]interface{}{
		"stake": 999999.0,
		"selections": []map[string]interface{}{
			{
				"market_id": testMarketID,
				"odd_id":    testOddIDs[0],
				"outcome":   "Home",
			},
		},
	}, s.userToken)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, body["error"].(string), "insufficient")
}

func (s *E2ESuite) Test24_PlaceBet_Unauthorized() {
	t := s.T()

	resp, _ := s.request("POST", "/api/bets", map[string]interface{}{
		"stake":      10.0,
		"selections": []map[string]interface{}{},
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func (s *E2ESuite) Test25_ListBets() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/bets", s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	bets := result["bets"].([]interface{})
	assert.GreaterOrEqual(t, len(bets), 1)
}

func (s *E2ESuite) Test26_GetBet() {
	t := s.T()
	require.NotEmpty(t, testBetID)

	resp, body := s.request("GET", "/api/bets/"+testBetID, nil, s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testBetID, body["id"])
	assert.Equal(t, "pending", body["status"])
	assert.NotNil(t, body["legs"])
}

func (s *E2ESuite) Test27_ListTransactions() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/transactions", s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	txs := result["transactions"].([]interface{})
	assert.GreaterOrEqual(t, len(txs), 1) // At least deposit + bet
}

func (s *E2ESuite) Test28_BalanceAfterBet() {
	t := s.T()

	resp, body := s.request("GET", "/api/auth/profile", nil, s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	user := body["user"].(map[string]interface{})
	balance := user["balance"].(float64)
	assert.Less(t, balance, 1000.0)
}

// ========================================================================
// 8. Admin Dashboard & Bet Management
// ========================================================================

func (s *E2ESuite) Test29_Admin_Stats() {
	t := s.T()

	resp, body := s.request("GET", "/api/admin/stats", nil, s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, body["total_users"].(float64), 1.0)
	assert.GreaterOrEqual(t, body["total_bets"].(float64), 1.0)
	assert.GreaterOrEqual(t, body["pending_bets"].(float64), 1.0)
	assert.Greater(t, body["total_staked"].(float64), 0.0)
}

func (s *E2ESuite) Test30_Admin_Stats_Unauthorized() {
	t := s.T()

	resp, _ := s.request("GET", "/api/admin/stats", nil, s.userToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func (s *E2ESuite) Test31_Admin_ListBets() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/admin/bets", s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	bets := result["bets"].([]interface{})
	assert.GreaterOrEqual(t, len(bets), 1)
}

func (s *E2ESuite) Test32_Admin_ListBets_ByStatus() {
	t := s.T()

	resp, data := s.requestArray("GET", "/api/admin/bets?status=pending", s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	bets := result["bets"].([]interface{})
	for _, b := range bets {
		bet := b.(map[string]interface{})
		assert.Equal(t, "pending", bet["status"])
	}
}

// ========================================================================
// 9. Admin Event Status Transitions
// ========================================================================

func (s *E2ESuite) Test33_Admin_UpdateEventStatus_ToLive() {
	t := s.T()

	resp, body := s.request("PATCH", "/api/admin/events/"+testEventID+"/status", map[string]string{
		"status": "live",
	}, s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, body)
}

func (s *E2ESuite) Test34_Admin_UpdateEventStatus_ToFinished() {
	t := s.T()

	resp, _ := s.request("PATCH", "/api/admin/events/"+testEventID+"/status", map[string]string{
		"status": "finished",
	}, s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ========================================================================
// 10. Settlement Flow
// ========================================================================

func (s *E2ESuite) Test35_Admin_SettleMarket() {
	t := s.T()

	// Get the first outcome from the odds
	_, data := s.requestArray("GET", "/api/events/"+testEventID+"/markets", "")
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	markets := result["markets"].([]interface{})
	market := markets[0].(map[string]interface{})
	odds := market["odds"].([]interface{})
	winningOutcome := odds[0].(map[string]interface{})["outcome"].(string)

	resp, body := s.request("POST", "/api/admin/settle", map[string]interface{}{
		"market_id":       testMarketID,
		"winning_outcome": winningOutcome,
	}, s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "settlement complete", body["message"])
	assert.GreaterOrEqual(t, body["settled_count"].(float64), 1.0)
}

func (s *E2ESuite) Test36_Admin_SettleMarket_AlreadySettled() {
	t := s.T()

	resp, body := s.request("POST", "/api/admin/settle", map[string]interface{}{
		"market_id":       testMarketID,
		"winning_outcome": "Home",
	}, s.adminToken)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, body["error"].(string), "already settled")
}

func (s *E2ESuite) Test37_BetStatus_AfterSettlement() {
	t := s.T()

	resp, body := s.request("GET", "/api/bets/"+testBetID, nil, s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	status := body["status"].(string)
	assert.Contains(t, []string{"won", "lost"}, status)

	if status == "won" {
		assert.Greater(t, body["actual_return"].(float64), 0.0)
	}
}

func (s *E2ESuite) Test38_Balance_AfterSettlement() {
	t := s.T()

	resp, body := s.request("GET", "/api/auth/profile", nil, s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	user := body["user"].(map[string]interface{})
	_ = user["balance"].(float64) // Just verify it's accessible
}

// ========================================================================
// 11. Odds Override
// ========================================================================

func (s *E2ESuite) Test39_Admin_CreateSecondEvent_ForOddsTest() {
	t := s.T()

	startTime := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)

	resp, body := s.request("POST", "/api/admin/events", map[string]interface{}{
		"sport_id": 1, "home_team": "Team C", "away_team": "Team D", "start_time": startTime,
	}, s.adminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	eventID := body["id"].(string)

	// Create market
	resp, body = s.request("POST", "/api/admin/markets", map[string]interface{}{
		"event_id": eventID, "market_type": "1x2", "name": "Match Result",
		"outcomes": []map[string]interface{}{
			{"outcome": "Home", "initial_odds": 1.80},
			{"outcome": "Draw", "initial_odds": 3.60},
			{"outcome": "Away", "initial_odds": 4.50},
		},
	}, s.adminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Get odds for override test
	marketID := body["id"].(string)
	_, data := s.requestArray("GET", "/api/markets/"+marketID+"/odds", "")
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	odds := result["odds"].([]interface{})
	require.NotEmpty(t, odds)

	oddID := odds[0].(map[string]interface{})["id"].(string)

	// Override odds
	resp, body = s.request("PUT", "/api/admin/odds/"+oddID, map[string]interface{}{
		"new_odds": 2.50,
	}, s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ========================================================================
// 12. Admin Event Deletion
// ========================================================================

func (s *E2ESuite) Test40_Admin_CreateAndDeleteEvent() {
	t := s.T()

	startTime := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)

	resp, body := s.request("POST", "/api/admin/events", map[string]interface{}{
		"sport_id": 1, "home_team": "Delete Me", "away_team": "Delete Test", "start_time": startTime,
	}, s.adminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	eventID := body["id"].(string)

	resp, _ = s.request("DELETE", "/api/admin/events/"+eventID, nil, s.adminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deleted
	resp, _ = s.request("GET", "/api/events/"+eventID, nil, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ========================================================================
// 13. Logout
// ========================================================================

func (s *E2ESuite) Test41_Auth_Logout() {
	t := s.T()

	resp, _ := s.request("POST", "/api/auth/logout", nil, s.userToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ========================================================================
// 14. Edge Cases
// ========================================================================

func (s *E2ESuite) Test42_GetEvent_NotFound() {
	t := s.T()

	resp, _ := s.request("GET", "/api/events/00000000-0000-0000-0000-000000000000", nil, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func (s *E2ESuite) Test43_GetBet_NotFound() {
	t := s.T()

	// Re-login since we logged out
	resp, body := s.request("POST", "/api/auth/login", map[string]string{
		"email": "bettor@betkz.test", "password": "Bettor1234!",
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	token := body["access_token"].(string)

	resp, _ = s.request("GET", "/api/bets/00000000-0000-0000-0000-000000000000", nil, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func (s *E2ESuite) Test44_Admin_InvalidDepositAmount() {
	t := s.T()

	resp, _ := s.request("POST", "/api/admin/users/"+s.userID+"/deposit", map[string]interface{}{}, s.adminToken)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

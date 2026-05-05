#!/usr/bin/env bash
# =============================================================================
# BetKZ Configuration Validation Script — Assignment 6
# Validates environment variables and service configuration BEFORE deployment.
# Prevents incidents like Assignment 4 (misconfigured Order Service backend URL).
# Usage: ./scripts/validate_config.sh
# =============================================================================

set -uo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BOLD='\033[1m'; NC='\033[0m'

PASS=0; FAIL=0; WARN=0

ok()   { echo -e "${GREEN}  ✓${NC} $*"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}  ✗${NC} $*"; FAIL=$((FAIL+1)); }
warn() { echo -e "${YELLOW}  !${NC} $*"; WARN=$((WARN+1)); }
section() { echo -e "\n${BOLD}── $* ──────────────────────────────────${NC}"; }

echo -e "${BOLD}============================================================${NC}"
echo -e "${BOLD}  BetKZ Pre-Deployment Configuration Validator${NC}"
echo -e "${BOLD}  Assignment 6 — Automation in SRE${NC}"
echo -e "${BOLD}============================================================${NC}"

# ── 1. Check .env file presence ──────────────────────────────────────────────
section "1. Environment Files"
ENV_FILE="deployments/.env"
if [[ -f "$ENV_FILE" ]]; then
  ok ".env file found at $ENV_FILE"
else
  warn ".env file missing — defaults will be used (OK for dev, NOT for prod)"
  ENV_FILE="deployments/.env.example"
fi

# Source env if present
if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  set -a; source "$ENV_FILE" 2>/dev/null || true; set +a
fi

# ── 2. Required environment variables ────────────────────────────────────────
section "2. Required Variables"
check_var() {
  local VAR="$1"
  local VAL="${!VAR:-}"
  if [[ -z "$VAL" ]]; then
    fail "$VAR is not set"
  else
    # Mask secrets
    case "$VAR" in
      *PASSWORD*|*SECRET*|*KEY*)
        ok "$VAR is set (***masked***)"
        ;;
      *)
        ok "$VAR = $VAL"
        ;;
    esac
  fi
}

check_var "DB_HOST"
check_var "DB_PORT"
check_var "DB_USER"
check_var "DB_PASSWORD"
check_var "DB_NAME"
check_var "REDIS_URL"
check_var "JWT_SECRET"

# ── 3. Insecure default detection ────────────────────────────────────────────
section "3. Security Checks"
DB_PASS="${DB_PASSWORD:-betkz_dev_pass}"
JWT_SEC="${JWT_SECRET:-change-me-in-production}"

if [[ "$DB_PASS" == "betkz_dev_pass" || "$DB_PASS" == "change_me_strong_password" ]]; then
  warn "DB_PASSWORD is set to a default/insecure value — change before production deploy"
fi
if [[ "$JWT_SEC" == "change-me-in-production" ]]; then
  warn "JWT_SECRET is set to the default placeholder — generate a real secret with: openssl rand -hex 32"
fi

# ── 4. Docker Compose validation ─────────────────────────────────────────────
section "4. Docker Compose Configuration"
COMPOSE_FILE="deployments/docker-compose.yml"
if [[ -f "$COMPOSE_FILE" ]]; then
  ok "docker-compose.yml found"
  # Check for restart policies
  if grep -q "restart: unless-stopped" "$COMPOSE_FILE"; then
    ok "restart: unless-stopped policy is present"
  else
    fail "No restart policy found — services won't auto-recover"
  fi
  # Check for health checks
  HC_COUNT=$(grep -c "healthcheck:" "$COMPOSE_FILE" || true)
  if [[ "$HC_COUNT" -ge 5 ]]; then
    ok "Health checks configured on $HC_COUNT services"
  else
    warn "Only $HC_COUNT health checks found — some services may not be monitored"
  fi
  # Check for the fault injection env variable being OFF
  INJECT_FAULT_VAL=$(grep "INJECT_FAULT" "$COMPOSE_FILE" | grep -v "#" | grep -o '"false"\|"true"' | tr -d '"' | head -1 || echo "false")
  if [[ "$INJECT_FAULT_VAL" == "false" ]]; then
    ok "INJECT_FAULT is set to 'false' — fault injection is disabled"
  else
    fail "INJECT_FAULT is not 'false' — ORDER SERVICE WILL FAIL! (this caused the Assignment 4 incident)"
  fi
else
  fail "docker-compose.yml not found at $COMPOSE_FILE"
fi

# ── 5. Prometheus configuration ──────────────────────────────────────────────
section "5. Monitoring Configuration"
PROM_FILE="deployments/prometheus/prometheus.yml"
ALERT_FILE="deployments/prometheus/alert.rules.yml"

if [[ -f "$PROM_FILE" ]]; then
  ok "prometheus.yml found"
  if grep -q "rule_files" "$PROM_FILE"; then
    ok "Alert rules are referenced in prometheus.yml"
  else
    warn "No rule_files section in prometheus.yml — alerting is not configured"
  fi
else
  fail "prometheus.yml not found"
fi

if [[ -f "$ALERT_FILE" ]]; then
  RULE_COUNT=$(grep -c "^      - alert:" "$ALERT_FILE" || true)
  ok "alert.rules.yml found — $RULE_COUNT alert rules defined"
else
  warn "alert.rules.yml not found — no alerting configured"
fi

# ── 6. Service backend URL validation ────────────────────────────────────────
section "6. Service Backend URL Validation"
# Check compose for any obviously wrong backend URLs (like the incident in Assignment 4)
if grep -q "backend-broken\|backend_wrong\|localhost:9999" "$COMPOSE_FILE" 2>/dev/null; then
  fail "CRITICAL: Found a broken/wrong backend URL in docker-compose.yml — this will cause service failures!"
else
  ok "No obviously broken backend URLs detected in docker-compose.yml"
fi

# Check for FAULTY_BACKEND_URL in environment
if grep -q "FAULTY_BACKEND_URL" "$COMPOSE_FILE" 2>/dev/null; then
  warn "FAULTY_BACKEND_URL is defined — ensure INJECT_FAULT=false to prevent activation"
fi

# ── 7. Docker presence ───────────────────────────────────────────────────────
section "7. Runtime Prerequisites"
if command -v docker &>/dev/null; then
  DOCKER_VERSION=$(docker --version 2>/dev/null | cut -d' ' -f3 || echo "unknown")
  ok "Docker is installed (${DOCKER_VERSION})"
else
  fail "Docker is not installed or not in PATH"
fi

if docker compose version &>/dev/null 2>&1; then
  ok "Docker Compose plugin is available"
elif command -v docker-compose &>/dev/null; then
  ok "docker-compose (standalone) is available"
else
  fail "Docker Compose is not available"
fi

# ── Summary ──────────────────────────────────────────────────────────────────
echo -e "\n${BOLD}============================================================${NC}"
echo -e "${BOLD}  VALIDATION SUMMARY${NC}"
echo -e "${BOLD}============================================================${NC}"
echo -e "  ${GREEN}Passed${NC}  : $PASS"
echo -e "  ${YELLOW}Warnings${NC}: $WARN"
echo -e "  ${RED}Failed${NC}  : $FAIL"
echo ""

if [[ "$FAIL" -gt 0 ]]; then
  echo -e "${RED}  ✗ VALIDATION FAILED — Do NOT deploy until issues are resolved.${NC}"
  exit 1
elif [[ "$WARN" -gt 0 ]]; then
  echo -e "${YELLOW}  ! Validation passed with warnings — review before production deploy.${NC}"
  exit 0
else
  echo -e "${GREEN}  ✓ All checks passed — safe to deploy.${NC}"
  echo -e "  Run: cd deployments && docker compose up -d"
  exit 0
fi

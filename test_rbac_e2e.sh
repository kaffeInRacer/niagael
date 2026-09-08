#!/usr/bin/env bash
set -euo pipefail

AUTH_URL="${AUTH_URL:-http://localhost:8085}"
PRODUCT_URL="${PRODUCT_URL:-http://localhost:8080}"
POLICY='{"service":"product","role":"staff","resource":"products","action":"delete"}'
MISSING_PRODUCT_ID="00000000-0000-0000-0000-000000000000"
ADMIN_COOKIE=$(mktemp /tmp/rbac-admin.XXXXXX)
STAFF_COOKIE=$(mktemp /tmp/rbac-staff.XXXXXX)
POLICY_ADDED=false

cleanup() {
  if [ "$POLICY_ADDED" = true ]; then
    curl -sS -o /dev/null -X DELETE "$AUTH_URL/admin/rbac/policies" \
      -b "$ADMIN_COOKIE" -H "Content-Type: application/json" -d "$POLICY" || true
  fi
  rm -f "$ADMIN_COOKIE" "$STAFF_COOKIE"
}
trap cleanup EXIT

request_status() {
  curl -sS -o /dev/null -w "%{http_code}" "$@"
}

wait_for_status() {
  local expected=$1
  shift
  local actual
  for _ in {1..60}; do
    actual=$(request_status "$@")
    if [ "$actual" = "$expected" ]; then
      printf '%s' "$actual"
      return 0
    fi
    sleep 0.25
  done
  printf '%s' "$actual"
  return 1
}

printf '%s\n' "=== E2E RBAC real-time policy reload ==="

admin_login=$(request_status -c "$ADMIN_COOKIE" -X POST "$AUTH_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@demo.com","password":"admin1234"}')
staff_login=$(request_status -c "$STAFF_COOKIE" -X POST "$AUTH_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"staff@demo.com","password":"staff1234"}')
[ "$admin_login" = "200" ] || { printf 'FAIL admin login: HTTP %s\n' "$admin_login"; exit 1; }
[ "$staff_login" = "200" ] || { printf 'FAIL staff login: HTTP %s\n' "$staff_login"; exit 1; }
printf '%s\n' "PASS login admin and staff"

# Force known baseline. A missing policy returns 404 and is already valid state.
baseline=$(request_status -X DELETE "$AUTH_URL/admin/rbac/policies" \
  -b "$ADMIN_COOKIE" -H "Content-Type: application/json" -d "$POLICY")
case "$baseline" in
  200|404) ;;
  *) printf 'FAIL baseline policy removal: HTTP %s\n' "$baseline"; exit 1 ;;
esac

if ! denied=$(wait_for_status 403 -X DELETE "$PRODUCT_URL/admin/products/$MISSING_PRODUCT_ID" -b "$STAFF_COOKIE"); then
  printf 'FAIL initial denial: expected 403, got %s\n' "$denied"
  exit 1
fi
printf '%s\n' "PASS staff delete denied: HTTP $denied"

added=$(request_status -X POST "$AUTH_URL/admin/rbac/policies" \
  -b "$ADMIN_COOKIE" -H "Content-Type: application/json" -d "$POLICY")
[ "$added" = "201" ] || { printf 'FAIL policy add: HTTP %s\n' "$added"; exit 1; }
POLICY_ADDED=true
printf '%s\n' "PASS policy added: HTTP $added"

# Missing product reaches handler only when authorization succeeds, producing 404.
if ! allowed=$(wait_for_status 404 -X DELETE "$PRODUCT_URL/admin/products/$MISSING_PRODUCT_ID" -b "$STAFF_COOKIE"); then
  printf 'FAIL policy propagation: expected 404, got %s\n' "$allowed"
  exit 1
fi
printf '%s\n' "PASS Kafka grant propagated: HTTP $allowed"

removed=$(request_status -X DELETE "$AUTH_URL/admin/rbac/policies" \
  -b "$ADMIN_COOKIE" -H "Content-Type: application/json" -d "$POLICY")
[ "$removed" = "200" ] || { printf 'FAIL policy removal: HTTP %s\n' "$removed"; exit 1; }
POLICY_ADDED=false
printf '%s\n' "PASS policy removed: HTTP $removed"

if ! denied_again=$(wait_for_status 403 -X DELETE "$PRODUCT_URL/admin/products/$MISSING_PRODUCT_ID" -b "$STAFF_COOKIE"); then
  printf 'FAIL revoke propagation: expected 403, got %s\n' "$denied_again"
  exit 1
fi
printf '%s\n' "PASS Kafka revoke propagated: HTTP $denied_again"
printf '%s\n' "=== ALL RBAC E2E TESTS PASSED ==="

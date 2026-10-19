#!/usr/bin/env bash
# Smoke test for a running stack (plan §11). Needs only bash, curl, grep and sed.
#
#   scripts/smoke.sh                          # through nginx: https://localhost
#   BASE_URL=https://localhost:8080 scripts/smoke.sh  # API directly (docker-compose.dev.yml)
#
# Exits 0 when every check passes, 1 otherwise.
set -uo pipefail

BASE_URL="${BASE_URL:-https://localhost}"
API="${BASE_URL%/}/api/v1"
WAIT_SECONDS="${WAIT_SECONDS:-60}"

failures=0
pass() { printf '  ok    %s\n' "$1"; }
fail() { printf '  FAIL  %s\n' "$1"; failures=$((failures + 1)); }

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# get NAME PATH [curl args...]: saves headers to $tmp/NAME.h, body to $tmp/NAME.b
# and prints the HTTP status. -k accepts self-signed certificates (nginx fallback, API).
get() {
  local name="$1" path="$2"
  shift 2
  curl -sSk -o "$tmp/$name.b" -D "$tmp/$name.h" -w '%{http_code}' "$@" "$API$path" 2>/dev/null || echo 000
}

header() { grep -i "^$2:" "$tmp/$1.h" | head -n1 | sed 's/^[^:]*: *//; s/\r$//'; }

# json_string NAME KEY: the first top-level-looking "KEY":"value" in the body.
json_string() { grep -o "\"$2\":\"[^\"]*\"" "$tmp/$1.b" | head -n1 | sed "s/^\"$2\":\"//; s/\"\$//"; }

echo "Smoke test against $API"

# 1. Wait for health.
deadline=$((SECONDS + WAIT_SECONDS))
until [ "$(get health /health)" = 200 ]; do
  if [ "$SECONDS" -ge "$deadline" ]; then
    fail "health did not return 200 within ${WAIT_SECONDS}s"
    exit 1
  fi
  sleep 2
done
if grep -q '"db":"ok"' "$tmp/health.b"; then pass "health: 200, db ok"; else fail "health body: $(cat "$tmp/health.b")"; fi

# 2. /main in English and Russian.
status="$(get main_en /main -H 'Accept-Language: en')"
name_en="$(json_string main_en fullName)"
if [ "$status" = 200 ] && [ "$(header main_en Content-Language)" = en ] && [ -n "$name_en" ]; then
  pass "main en: fullName \"$name_en\""
else
  fail "main en: status $status, Content-Language '$(header main_en Content-Language)'"
fi

status="$(get main_ru /main -H 'Accept-Language: ru-RU,ru;q=0.9')"
name_ru="$(json_string main_ru fullName)"
if [ "$status" = 200 ] && [ "$(header main_ru Content-Language)" = ru ] && [ -n "$name_ru" ] && [ "$name_ru" != "$name_en" ]; then
  pass "main ru: fullName \"$name_ru\""
else
  fail "main ru: status $status, Content-Language '$(header main_ru Content-Language)', fullName '$name_ru'"
fi
if grep -qi '^Vary:.*Accept-Language' "$tmp/main_ru.h"; then # Vary may span several header lines
  pass "main: Vary includes Accept-Language"
else
  fail "main: Vary lacks Accept-Language"
fi

# 3. Responses are never cached: data comes fresh from the DB on every request.
cc="$(header main_ru Cache-Control)"
if [ "$cc" = no-store ]; then pass "main: Cache-Control no-store"; else fail "main: Cache-Control '$cc', want no-store"; fi

# 4. /cv and /travel.
status="$(get cv /cv)"
if [ "$status" = 200 ] && grep -q '"experience":\[' "$tmp/cv.b"; then pass "cv: 200 with experience"; else fail "cv: status $status"; fi

status="$(get travel /travel -H 'Accept-Language: ru')"
if [ "$status" = 200 ] && grep -q '"stats":{"countries":[0-9]' "$tmp/travel.b"; then
  pass "travel: 200 with stats $(grep -o '"stats":{[^}]*}' "$tmp/travel.b" | sed 's/"stats"://')"
else
  fail "travel: status $status, no stats"
fi

# 5. Errors use the JSON error format.
status="$(get missing /does-not-exist)"
if [ "$status" = 404 ] && grep -q '"code":"not_found"' "$tmp/missing.b"; then pass "unknown path: 404 not_found"; else fail "unknown path: status $status"; fi
status="$(get post /main -X POST)"
if [ "$status" = 405 ] && grep -q '"code":"method_not_allowed"' "$tmp/post.b"; then pass "POST: 405 method_not_allowed"; else fail "POST: status $status"; fi

if [ "$failures" -gt 0 ]; then
  echo "$failures check(s) failed"
  exit 1
fi
echo "All checks passed"

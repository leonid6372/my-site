#!/bin/sh
# Runs from the nginx image's /docker-entrypoint.d before nginx starts.
#
# Points /etc/nginx/certs/{fullchain,privkey}.pem at:
#   1. the Let's Encrypt certificate for $DOMAIN, once certbot has issued it;
#   2. otherwise a self-signed certificate, so nginx can start (and serve the
#      ACME challenge on port 80 that Let's Encrypt needs to issue the real one).
# A background loop re-checks every 5 min and reloads nginx when the
# certificate changed, which picks up the first issue and every renewal.
set -eu

CERTS=/etc/nginx/certs
SELF="$CERTS/self-signed"
mkdir -p "$CERTS" "$SELF"

link_certs() {
  le="/etc/letsencrypt/live/${DOMAIN:-}"
  if [ -n "${DOMAIN:-}" ] && [ -s "$le/fullchain.pem" ] && [ -s "$le/privkey.pem" ]; then
    src="$le"
  else
    if [ ! -s "$SELF/fullchain.pem" ]; then
      openssl req -x509 -nodes -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -days 365 \
        -subj "/CN=${DOMAIN:-localhost}" \
        -addext "subjectAltName=DNS:${DOMAIN:-localhost},DNS:localhost,IP:127.0.0.1" \
        -keyout "$SELF/privkey.pem" -out "$SELF/fullchain.pem" 2>/dev/null
    fi
    src="$SELF"
  fi
  ln -sf "$src/fullchain.pem" "$CERTS/fullchain.pem"
  ln -sf "$src/privkey.pem" "$CERTS/privkey.pem"
  echo "tls: using $src"
}

fingerprint() { cksum "$CERTS/fullchain.pem" 2>/dev/null || true; }

link_certs
last="$(fingerprint)"

(
  while sleep 300; do
    link_certs >/dev/null
    now="$(fingerprint)"
    if [ "$now" != "$last" ]; then
      echo "tls: certificate changed, reloading nginx"
      nginx -s reload
      last="$now"
    fi
  done
) &

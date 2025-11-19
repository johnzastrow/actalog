#!/usr/bin/env bash
set -euo pipefail

# start-frontend.sh
# Guided script to configure and start the ActaLog frontend (dev or preview)
# Usage: ./scripts/start-frontend.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$ROOT_DIR/web"

echo "\nActaLog frontend starter — guided mode"

read -r -p "Run in (d)ev or (p)review (production) mode? [d/p]: " MODE
MODE=${MODE:-d}

if [[ "$MODE" != "d" && "$MODE" != "p" ]]; then
  echo "Invalid choice. Use 'd' for dev or 'p' for preview." >&2
  exit 1
fi

read -r -p "Will you access the frontend on this machine's localhost or a domain name? [l=localhost / D=domain]: " HOST_TYPE
HOST_TYPE=${HOST_TYPE:-l}

if [[ "$HOST_TYPE" == "l" || "$HOST_TYPE" == "L" ]]; then
  HOSTNAME="localhost"
  echo "Detected: localhost environment"
  read -r -p "Expose to LAN (so other devices can access)? [y/N]: " EXPOSE
  EXPOSE=${EXPOSE:-N}
  if [[ "$EXPOSE" =~ ^[Yy]$ ]]; then
    HOST_ARG="--host"
  else
    HOST_ARG=""
  fi

  read -r -p "Use HTTPS locally (requires mkcert or custom cert)? [y/N]: " USE_HTTPS
  USE_HTTPS=${USE_HTTPS:-N}
  if [[ "$USE_HTTPS" =~ ^[Yy]$ ]]; then
    HTTPS_FLAG="--https"
    echo "If you want trusted certs, mkcert can create locally-trusted certs (https://github.com/FiloSottile/mkcert)."
    if command -v mkcert >/dev/null 2>&1; then
      read -r -p "mkcert is installed — generate certs for '$HOSTNAME' now? [y/N]: " MKCERT_YES
      MKCERT_YES=${MKCERT_YES:-N}
      if [[ "$MKCERT_YES" =~ ^[Yy]$ ]]; then
        CERT_DIR="$WEB_DIR/certs"
        mkdir -p "$CERT_DIR"
        echo "Running: mkcert -install (may prompt for permission)..."
        mkcert -install || true
        CERT_PEM="$CERT_DIR/$HOSTNAME.pem"
        CERT_KEY="$CERT_DIR/$HOSTNAME-key.pem"
        echo "Generating certs for: $HOSTNAME (and localhost/127.0.0.1)..."
        mkcert -cert-file "$CERT_PEM" -key-file "$CERT_KEY" "$HOSTNAME" localhost 127.0.0.1
        echo "Created cert files:"
        echo "  $CERT_PEM"
        echo "  $CERT_KEY"
        echo "\nNext steps (one-time):"
        echo "- Keep these files under 'web/certs' (already created)."
        echo "- Update your 'web/vite.config.js' to point Vite's server.https to these files (example below)."
        echo "\nPaste this snippet into your 'web/vite.config.js' (add 'import fs from \"fs\"'):\n"
        cat <<EOF
server: {
  host: '$HOSTNAME',
  https: {
    key: fs.readFileSync('certs/$HOSTNAME-key.pem'),
    cert: fs.readFileSync('certs/$HOSTNAME.pem')
  }
}
EOF
        echo "\nOr run preview with flags (does not require editing config):"
        echo "  npm run preview -- --https --host $HOSTNAME"
      else
        echo "Skipping mkcert generation — you'll need to provide certs and update Vite config if you want HTTPS."
      fi
    else
      echo "mkcert not found. Install it from https://github.com/FiloSottile/mkcert and re-run this script to auto-generate certs."
      echo "You can still run preview with '--https' if you provide certs in Vite config."
    fi
  else
    HTTPS_FLAG=""
  fi

else
  read -r -p "Enter the domain you'll use (e.g. actalog.example.com): " HOSTNAME
  if [[ -z "$HOSTNAME" ]]; then
    echo "A domain name is required for domain mode." >&2
    exit 1
  fi
  echo "Domain mode: $HOSTNAME"

  read -r -p "Will you use HTTPS for this domain? [Y/n]: " USE_HTTPS
  USE_HTTPS=${USE_HTTPS:-Y}
  if [[ "$USE_HTTPS" =~ ^[Yy]$ ]]; then
    HTTPS_FLAG="--https"
    echo "Ensure DNS A/AAA record points to this machine and that you have certs available (mkcert for local dev or real certs in production)."
  else
    HTTPS_FLAG=""
  fi

  # Give guidance about hosts file if it's local testing with domain
  echo "If you're testing a domain locally, add an entry to /etc/hosts (or C:\\Windows\\System32\\drivers\\etc\\hosts on Windows):"
  echo "  <your-machine-ip>  $HOSTNAME"
fi

echo "\nSummary:"
echo " Mode:       $([[ $MODE == 'd' ]] && echo 'dev' || echo 'preview')"
echo " Hostname:   $HOSTNAME"
echo " HTTPS flag: ${HTTPS_FLAG:-none}"

cd "$WEB_DIR"

if [[ "$MODE" == "d" ]]; then
  echo "\nStarting frontend in development mode..."
  # Use npm script (vite). Provide host/https flags directly to the vite CLI.
  # `npm run dev -- --host` passes flags through to vite.
  if [[ -n "$HOST_ARG" ]]; then
    echo "Running: npm run dev -- $HOST_ARG $HOSTNAME $HTTPS_FLAG"
    exec npm run dev -- $HOST_ARG $HOSTNAME $HTTPS_FLAG
  else
    echo "Running: npm run dev $HTTPS_FLAG"
    exec npm run dev -- $HTTPS_FLAG
  fi
else
  echo "\nBuilding and starting preview (production-like) server..."
  echo "Running: npm run build && npm run preview ${HTTPS_FLAG} --host $HOSTNAME"
  # preview accepts --host and --https
  npm run build
  if [[ -n "$HTTPS_FLAG" ]]; then
    exec npm run preview -- --https --host $HOSTNAME
  else
    exec npm run preview -- --host $HOSTNAME
  fi
fi

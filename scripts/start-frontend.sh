#!/usr/bin/env bash
set -euo pipefail

# start-frontend.sh
# Guided script to configure and start the ActaLog frontend (dev or preview)
# Usage: ./scripts/start-frontend.sh

# Script version - increment when making changes
SCRIPT_VERSION="1.2.0"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$ROOT_DIR/web"

# Defaults to avoid 'unbound variable' with 'set -u'
HOST_ARG=""
HTTPS_FLAG=""

echo ""
echo "ActaLog Frontend Starter v${SCRIPT_VERSION}"
echo "========================================="
echo ""

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
        echo ""
        echo "Next steps (one-time):"
        echo "- Keep these files under 'web/certs' (already created)."
        echo "- Update your 'web/vite.config.js' to point Vite's server.https to these files (example below)."
        echo ""
        echo "Paste this snippet into your 'web/vite.config.js' (add 'import fs from \"fs\"'):"
        echo ""
        cat <<EOF
server: {
  host: '$HOSTNAME',
  https: {
    key: fs.readFileSync('certs/$HOSTNAME-key.pem'),
    cert: fs.readFileSync('certs/$HOSTNAME.pem')
  }
}
EOF
        echo ""
        echo "Or run preview with flags (does not require editing config):"
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

  # Verify DNS resolution
  echo "\nVerifying DNS configuration for $HOSTNAME..."
  RESOLVED_IP=""
  if command -v dig >/dev/null 2>&1; then
    # Use dig if available (more reliable)
    RESOLVED_IP=$(dig +short "$HOSTNAME" A | head -n 1)
  elif command -v nslookup >/dev/null 2>&1; then
    # Fall back to nslookup
    RESOLVED_IP=$(nslookup "$HOSTNAME" 2>/dev/null | grep -A 1 "Name:" | tail -n 1 | awk '{print $2}')
  elif command -v host >/dev/null 2>&1; then
    # Fall back to host command
    RESOLVED_IP=$(host "$HOSTNAME" 2>/dev/null | grep "has address" | head -n 1 | awk '{print $NF}')
  fi

  if [[ -n "$RESOLVED_IP" ]]; then
    echo "✓ DNS resolved: $HOSTNAME → $RESOLVED_IP"

    # Get local machine's IP addresses
    LOCAL_IPS=""
    if command -v hostname >/dev/null 2>&1; then
      LOCAL_IPS=$(hostname -I 2>/dev/null || hostname -i 2>/dev/null || echo "")
    fi

    # Check if resolved IP matches any local IP
    IP_MATCHES=false
    if [[ -n "$LOCAL_IPS" ]]; then
      for LOCAL_IP in $LOCAL_IPS; do
        if [[ "$RESOLVED_IP" == "$LOCAL_IP" ]]; then
          IP_MATCHES=true
          echo "✓ DNS points to this machine ($LOCAL_IP)"
          break
        fi
      done

      if [[ "$IP_MATCHES" == false ]]; then
        echo "⚠ WARNING: DNS does not point to this machine!"
        echo "  Resolved IP: $RESOLVED_IP"
        echo "  Local IPs:   $LOCAL_IPS"
        echo "\nThis may cause issues unless:"
        echo "  - You're behind a NAT/router doing port forwarding"
        echo "  - You're testing with /etc/hosts override"
        echo "  - The domain points to a load balancer that forwards here"
      fi
    else
      echo "Could not determine local IP addresses for comparison."
    fi

    read -r -p "\nIs the resolved IP ($RESOLVED_IP) correct? [Y/n]: " IP_CORRECT
    IP_CORRECT=${IP_CORRECT:-Y}
    if [[ ! "$IP_CORRECT" =~ ^[Yy]$ ]]; then
      echo "\n✗ DNS configuration issue detected!"
      echo "Please fix your DNS records before continuing:"
      echo "  1. Add/update an A record for '$HOSTNAME' pointing to your server's public IP"
      echo "  2. Wait for DNS propagation (can take 5 minutes to 48 hours)"
      echo "  3. Verify with: dig $HOSTNAME"
      read -r -p "\nContinue anyway? [y/N]: " CONTINUE_ANYWAY
      CONTINUE_ANYWAY=${CONTINUE_ANYWAY:-N}
      if [[ ! "$CONTINUE_ANYWAY" =~ ^[Yy]$ ]]; then
        echo "Exiting. Fix DNS configuration and re-run this script."
        exit 1
      fi
    fi
  else
    echo "✗ WARNING: Could not resolve DNS for '$HOSTNAME'"
    echo "This domain does not appear to have a valid A record."
    echo "\nPossible issues:"
    echo "  - Domain/subdomain does not exist"
    echo "  - DNS not yet propagated"
    echo "  - DNS server issue"
    echo "\nTo fix:"
    echo "  1. Add an A record for '$HOSTNAME' in your DNS provider"
    echo "  2. Point it to this server's public IP address"
    echo "  3. Wait for DNS propagation (typically 5-30 minutes)"
    echo "  4. Verify with: dig $HOSTNAME"

    read -r -p "\nContinue anyway (for local testing with /etc/hosts)? [y/N]: " CONTINUE_ANYWAY
    CONTINUE_ANYWAY=${CONTINUE_ANYWAY:-N}
    if [[ ! "$CONTINUE_ANYWAY" =~ ^[Yy]$ ]]; then
      echo "Exiting. Configure DNS and re-run this script."
      exit 1
    fi
  fi

  # Ask if using a reverse proxy (Caddy)
  read -r -p "Will you use a reverse proxy like Caddy to handle HTTPS? [Y/n]: " USE_PROXY
  USE_PROXY=${USE_PROXY:-Y}

  if [[ "$USE_PROXY" =~ ^[Yy]$ ]]; then
    echo "Reverse proxy mode: The frontend will run HTTP-only (no HTTPS flag needed)."
    echo "Your reverse proxy (e.g., Caddy) will handle HTTPS and forward requests to the frontend."
    echo "\nMake sure:"
    echo "  - DNS A/AAAA record points to this machine"
    echo "  - Reverse proxy is configured to proxy to localhost:3000 (or your configured port)"
    echo "  - Reverse proxy handles SSL/TLS certificates (Caddy does this automatically via Let's Encrypt)"
    HTTPS_FLAG=""
  else
    read -r -p "Will you use HTTPS directly on the frontend? [Y/n]: " USE_HTTPS
    USE_HTTPS=${USE_HTTPS:-Y}
    if [[ "$USE_HTTPS" =~ ^[Yy]$ ]]; then
      HTTPS_FLAG="--https"
      echo "Ensure DNS A/AAAA record points to this machine and that you have certs available (mkcert for local dev or real certs in production)."
    else
      HTTPS_FLAG=""
    fi
  fi

  # Give guidance about hosts file if it's local testing with domain
  echo "\nIf you're testing a domain locally, add an entry to /etc/hosts (or C:\\Windows\\System32\\drivers\\etc\\hosts on Windows):"
  echo "  <your-machine-ip>  $HOSTNAME"
fi

echo "\nSummary:"
echo " Mode:       $([[ $MODE == 'd' ]] && echo 'dev' || echo 'preview')"
echo " Hostname:   $HOSTNAME"
echo " HTTPS flag: ${HTTPS_FLAG:-none}"

# Check if port 3000 is already in use
PORT=3000
PORT_IN_USE=false
PROCESS_INFO=""

if command -v lsof >/dev/null 2>&1; then
  PROCESS_INFO=$(lsof -ti:$PORT 2>/dev/null | head -n 1 || true)
elif command -v ss >/dev/null 2>&1; then
  PROCESS_INFO=$(ss -tlnp 2>/dev/null | grep ":$PORT " | head -n 1 || true)
elif command -v netstat >/dev/null 2>&1; then
  PROCESS_INFO=$(netstat -tlnp 2>/dev/null | grep ":$PORT " | head -n 1 || true)
fi

if [[ -n "$PROCESS_INFO" ]]; then
  PORT_IN_USE=true
  echo "\n⚠ WARNING: Port $PORT is already in use!"

  # Try to get more detailed process information
  if command -v lsof >/dev/null 2>&1; then
    PID=$(lsof -ti:$PORT 2>/dev/null | head -n 1 || true)
    if [[ -n "$PID" ]]; then
      PROC_NAME=$(ps -p "$PID" -o comm= 2>/dev/null || echo "unknown")
      PROC_CMD=$(ps -p "$PID" -o args= 2>/dev/null || echo "unknown")
      echo "  Process ID:   $PID"
      echo "  Process name: $PROC_NAME"
      echo "  Command:      $PROC_CMD"
    else
      echo "  $PROCESS_INFO"
    fi
  else
    echo "  $PROCESS_INFO"
  fi

  echo "\nOptions:"
  echo "  1) Stop the existing process and start ActaLog on port $PORT"
  echo "  2) Start ActaLog on a different port"
  echo "  3) Cancel and exit"

  read -r -p "\nChoose an option [1/2/3]: " PORT_CHOICE
  PORT_CHOICE=${PORT_CHOICE:-3}

  case "$PORT_CHOICE" in
    1)
      echo "\nAttempting to stop process on port $PORT..."
      if command -v lsof >/dev/null 2>&1; then
        PID=$(lsof -ti:$PORT 2>/dev/null | head -n 1 || true)
        if [[ -n "$PID" ]]; then
          echo "Killing process $PID..."
          kill "$PID" 2>/dev/null || kill -9 "$PID" 2>/dev/null || {
            echo "✗ Failed to kill process. You may need sudo privileges."
            read -r -p "Try with sudo? [y/N]: " USE_SUDO
            USE_SUDO=${USE_SUDO:-N}
            if [[ "$USE_SUDO" =~ ^[Yy]$ ]]; then
              sudo kill "$PID" 2>/dev/null || sudo kill -9 "$PID" 2>/dev/null || {
                echo "✗ Failed to kill process even with sudo. Exiting."
                exit 1
              }
            else
              echo "Exiting."
              exit 1
            fi
          }
          # Wait for port to be released
          sleep 2
          echo "✓ Process stopped. Proceeding with port $PORT..."
        fi
      else
        echo "✗ Cannot automatically kill process (lsof not available)."
        echo "Please manually stop the process using port $PORT and re-run this script."
        exit 1
      fi
      ;;
    2)
      echo "\nFinding an available port..."
      # Find next available port starting from 3001
      for TEST_PORT in {3001..3010}; do
        if ! lsof -ti:$TEST_PORT >/dev/null 2>&1 && ! ss -tln 2>/dev/null | grep -q ":$TEST_PORT "; then
          PORT=$TEST_PORT
          echo "✓ Found available port: $PORT"
          break
        fi
      done

      if [[ "$PORT" == "3000" ]]; then
        echo "✗ Could not find an available port in range 3001-3010"
        exit 1
      fi

      echo "\n📌 IMPORTANT: ActaLog will start on port $PORT"
      echo "   Access it at: http://$HOSTNAME:$PORT"
      if [[ "$USE_PROXY" =~ ^[Yy]$ ]]; then
        echo "\n   ⚠ Remember to update your reverse proxy configuration!"
        echo "   Change 'reverse_proxy localhost:3000' to 'reverse_proxy localhost:$PORT'"
        echo "   in your Caddyfile or nginx config."
      fi
      ;;
    3)
      echo "Exiting."
      exit 0
      ;;
    *)
      echo "Invalid choice. Exiting."
      exit 1
      ;;
  esac
fi

cd "$WEB_DIR"

# Show final configuration
echo ""
echo "========================================="
echo "Starting ActaLog Frontend v${SCRIPT_VERSION}"
echo "========================================="
echo " Port:     $PORT"
echo " Mode:     $([[ $MODE == 'd' ]] && echo 'Development' || echo 'Preview (Production-like)')"
echo " Hostname: $HOSTNAME"
if [[ "$PORT" != "3000" ]]; then
  echo ""
  echo "⚠ Using non-standard port: $PORT"
  echo "   Access URL: http://$HOSTNAME:$PORT"
fi
echo "========================================="
echo ""

# Export environment variables for Vite configuration
export VITE_DEV_HOST="$HOSTNAME"
export VITE_DEV_PORT="$PORT"

# Set HTTPS flag for Vite
if [[ -n "$HTTPS_FLAG" ]]; then
  export VITE_USE_HTTPS="true"
else
  export VITE_USE_HTTPS="false"
fi

# Determine bind address for Vite server
# When using reverse proxy, bind to 0.0.0.0 (all interfaces), not the domain name
if [[ "$HOST_TYPE" == "l" || "$HOST_TYPE" == "L" ]]; then
  BIND_HOST="$HOSTNAME"  # localhost or other host from config
else
  # Domain mode with reverse proxy: bind to 0.0.0.0
  if [[ "${USE_PROXY:-N}" =~ ^[Yy]$ ]]; then
    BIND_HOST="0.0.0.0"
  else
    BIND_HOST="$HOSTNAME"
  fi
fi

# Construct deployment URL for PWA manifest and build configuration
if [[ "$HOST_TYPE" == "l" || "$HOST_TYPE" == "L" ]]; then
  # Localhost mode - always include port
  if [[ -n "$HTTPS_FLAG" ]]; then
    export VITE_DEPLOYMENT_URL="https://$HOSTNAME:$PORT"
  else
    export VITE_DEPLOYMENT_URL="http://$HOSTNAME:$PORT"
  fi
else
  # Domain mode
  if [[ "${USE_PROXY:-N}" =~ ^[Yy]$ ]]; then
    # Using reverse proxy - don't include port in public URL
    # (proxy listens on 80/443 and forwards to localhost:$PORT)
    export VITE_DEPLOYMENT_URL="https://$HOSTNAME"
  else
    # Direct access - include port in URL
    if [[ -n "$HTTPS_FLAG" ]]; then
      export VITE_DEPLOYMENT_URL="https://$HOSTNAME:$PORT"
    else
      export VITE_DEPLOYMENT_URL="http://$HOSTNAME:$PORT"
    fi
  fi
fi

if [[ "$MODE" == "d" ]]; then
  # Use npm script (vite). Provide host/https flags directly to the vite CLI.
  # `npm run dev -- --host` passes flags through to vite.
  if [[ -n "$HOST_ARG" ]]; then
    echo "Running: npm run dev -- $HOST_ARG --port $PORT $HTTPS_FLAG"
    exec npm run dev -- $HOST_ARG --port $PORT $HTTPS_FLAG
  else
    echo "Running: npm run dev -- --port $PORT $HTTPS_FLAG"
    exec npm run dev -- --port $PORT $HTTPS_FLAG
  fi
else
  echo ""
  echo "Building and starting preview (production-like) server on port $PORT..."
  # Note: vite preview does NOT accept --https as a CLI flag
  # HTTPS is controlled via vite.config.js server.https setting
  # If certs exist in web/certs/, the preview server will use HTTPS automatically
  if [[ -n "$HTTPS_FLAG" ]]; then
    echo "Note: HTTPS for preview mode is configured via vite.config.js (not CLI flags)"
    echo "Ensure cert files exist in web/certs/ for HTTPS support"
    echo "Running: npm run build && npm run preview -- --host $BIND_HOST --port $PORT"
  else
    echo "Running: npm run build && npm run preview -- --host $BIND_HOST --port $PORT"
  fi
  npm run build
  exec npm run preview -- --host $BIND_HOST --port $PORT
fi

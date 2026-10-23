#!/bin/bash

# Start ScraperGo (Intel/amd64) and its Cloudflare Tunnel on port 8083.
# Leave the sailing and stock applications and tunnels untouched.
#
# pgrep -fl 'scrapergo-darwin-amd64|cloudflared'
# curl -fsS http://localhost:8083/

HOME=/Users/richardmauri
APP_DIR="$HOME/projects/scrapergo"
APP="$APP_DIR/dist/scrapergo-darwin-amd64"
TOKEN="$APP_DIR/.cloudflared/token"
CLOUDFLARED="$HOME/Applications/cloudflared"

cd "$APP_DIR" || exit 1

for file in "$APP" "$TOKEN" "$CLOUDFLARED"; do
    if [[ ! -f "$file" ]]; then
        echo "Missing required file: $file" >&2
        exit 1
    fi
done

# Show the version embedded in the built executable before replacing the running app.
"$APP" --version || exit 1

# Stop only the previous ScraperGo application.
pkill -x scrapergo-darwin-amd64 2>/dev/null || true

# Stop only the cloudflared process associated with this token file.
ps -axo pid=,command= | while read -r pid command; do
    if [[ "$command" == *"cloudflared tunnel run"* &&
          "$command" == *"$TOKEN"* ]]; then
        kill "$pid" 2>/dev/null || true
    fi
done

# Start ScraperGo on localhost port 8083.
nohup "$APP" -port 8083 \
    > "$APP_DIR/scrapergo.log" 2>&1 < /dev/null &

# Start only ScraperGo's Cloudflare Tunnel.
nohup "$CLOUDFLARED" tunnel run \
    --token-file "$TOKEN" \
    > "$APP_DIR/.cloudflared/tunnel.log" 2>&1 < /dev/null &

disown -a

echo "Started ScraperGo on port 8083 and its Cloudflare Tunnel."
"$APP" --version

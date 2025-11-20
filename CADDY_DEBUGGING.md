# Caddy Debugging Guide for al.fluidgrid.site

## Changes Made to Caddyfile

The Caddyfile has been updated with comprehensive debugging capabilities:

### 1. **Global Debug Mode**
```caddy
{
    debug                    # Enable verbose debug logging
    admin 0.0.0.0:2019      # Admin API accessible for diagnostics
}
```

### 2. **Enhanced Logging**
- **Log Level:** Changed from default to `DEBUG` for both sites
- **Log Location:**
  - ActaLog: `/var/log/caddy/al.access.log`
  - Mealie: `/var/log/caddy/recipe.access.log`
- **Format:** Human-readable console format

### 3. **Debug Response Headers**
Both sites now include debug headers in responses:
- `X-Debug-Proxy`: Confirms Caddy is proxying
- `X-Debug-Backend`: Shows which backend port is being used

### 4. **Diagnostic Endpoints**
- `https://al.fluidgrid.site/caddy-test` - Returns timestamp to confirm Caddy is serving
- `https://recipe.fluidgrid.site/caddy-test` - Returns timestamp to confirm Caddy is serving
- `http://your-server-ip:80` - Catch-all that shows requested hostname

### 5. **HTTP Port 80 Catch-All**
Added a catch-all listener on port 80 that will show what hostname was requested. This helps diagnose DNS/routing issues.

---

## Debugging Steps

### Step 1: Validate Caddyfile Syntax
```bash
# Validate the Caddyfile without starting Caddy
caddy validate --config /home/jcz/Github/actionlog/Caddyfile
```

### Step 2: Check Caddy is Running
```bash
# Check if Caddy is running
systemctl status caddy

# Or if running manually:
ps aux | grep caddy
```

### Step 3: Restart Caddy with New Config
```bash
# If using systemd:
sudo systemctl restart caddy

# Or reload without dropping connections:
sudo systemctl reload caddy

# If running manually:
caddy stop
caddy run --config /home/jcz/Github/actionlog/Caddyfile
```

### Step 4: Check Caddy Logs
```bash
# View Caddy's main log (systemd)
sudo journalctl -u caddy -f

# View ActaLog access logs
sudo tail -f /var/log/caddy/al.access.log

# View Mealie access logs
sudo tail -f /var/log/caddy/recipe.access.log
```

### Step 5: Test Diagnostic Endpoints

#### Test from Server Itself
```bash
# Test ActaLog diagnostic endpoint
curl -v http://localhost/caddy-test -H "Host: al.fluidgrid.site"

# Test if frontend is running
curl -v http://localhost:3000

# Test if backend is running
curl -v http://localhost:8080/health
```

#### Test from External Client
```bash
# Test diagnostic endpoint
curl -v https://al.fluidgrid.site/caddy-test

# Check response headers (should include X-Debug-Proxy and X-Debug-Backend)
curl -I https://al.fluidgrid.site/

# Test HTTP port 80 catch-all
curl -v http://your-server-ip/
```

### Step 6: Check DNS Resolution
```bash
# Verify DNS is resolving correctly
dig al.fluidgrid.site
nslookup al.fluidgrid.site

# From your local machine, verify DNS
ping al.fluidgrid.site
```

### Step 7: Check Firewall
```bash
# Ensure ports 80 and 443 are open
sudo ufw status

# If needed, allow them:
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

### Step 8: Verify Backend Services
```bash
# Check if ActaLog frontend is running on port 3000
netstat -tlnp | grep :3000
# Or
ss -tlnp | grep :3000

# Check if ActaLog backend is running on port 8080
netstat -tlnp | grep :8080

# Check if Mealie is running on port 9925
netstat -tlnp | grep :9925
```

### Step 9: Use Caddy Admin API
```bash
# Get current config (requires debug mode enabled)
curl http://localhost:2019/config/ | jq

# Get current adapters
curl http://localhost:2019/config/apps/http/servers
```

### Step 10: Check Certificate Status
```bash
# Check if Caddy obtained SSL certificates
sudo caddy list-certificates

# Or check certificate storage
ls -la /var/lib/caddy/.local/share/caddy/certificates/
```

---

## Common Issues & Solutions

### Issue 1: "Connection Refused"
**Symptoms:** Browser shows "Connection refused" or "Unable to connect"

**Possible Causes:**
1. Caddy is not running
2. Firewall blocking ports 80/443
3. DNS not pointing to server

**Solutions:**
```bash
# Check Caddy status
systemctl status caddy

# Check firewall
sudo ufw status

# Verify DNS
dig al.fluidgrid.site
```

### Issue 2: "502 Bad Gateway"
**Symptoms:** Caddy is running but backend not responding

**Possible Causes:**
1. Frontend (port 3000) not running
2. Backend (port 8080) not running
3. Services listening on wrong interface

**Solutions:**
```bash
# Check if frontend is running
curl http://localhost:3000

# Check if backend is running
curl http://localhost:8080/health

# Start frontend if needed
cd /home/jcz/Github/actionlog/web
npm run dev

# Start backend if needed
cd /home/jcz/Github/actionlog
./bin/actalog
```

### Issue 3: "Certificate Error"
**Symptoms:** Browser shows SSL/TLS certificate error

**Possible Causes:**
1. Caddy couldn't obtain Let's Encrypt certificate
2. DNS not propagated yet
3. Port 80 blocked (needed for ACME challenge)

**Solutions:**
```bash
# Check Caddy logs for certificate errors
journalctl -u caddy | grep -i cert

# Ensure port 80 is accessible from internet
# Let's Encrypt uses HTTP-01 challenge on port 80

# Force certificate renewal
caddy reload --force
```

### Issue 4: "Site Not Loading"
**Symptoms:** Caddy running, no errors, but site doesn't load

**Possible Causes:**
1. DNS not pointing to server
2. Wrong server IP in DNS
3. Requests going to different server

**Solutions:**
```bash
# Verify DNS is pointing to your server IP
dig al.fluidgrid.site +short

# Check what IP your server has
curl ifconfig.me

# Test from server itself
curl -v http://localhost -H "Host: al.fluidgrid.site"
```

---

## Expected Debug Output

When everything is working, you should see:

### 1. Response Headers
```
X-Debug-Proxy: ActaLog-Caddy
X-Debug-Backend: localhost:3000
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

### 2. Diagnostic Endpoint Response
```
Caddy is serving al.fluidgrid.site - 2025-11-20 18:30:45
```

### 3. Access Log Entry (DEBUG level)
```
{"level":"debug","ts":1700511045.123,"logger":"http.handlers.reverse_proxy","msg":"proxying request","request":{"remote_addr":"1.2.3.4:12345","proto":"HTTP/2.0","method":"GET","host":"al.fluidgrid.site","uri":"/","headers":{"User-Agent":["Mozilla/5.0..."]}}}
```

### 4. Caddy Status
```bash
$ systemctl status caddy
● caddy.service - Caddy
     Loaded: loaded
     Active: active (running)
```

---

## Quick Troubleshooting Commands

```bash
# One-liner to check everything
echo "=== Caddy Status ===" && systemctl status caddy --no-pager && \
echo -e "\n=== Ports ===" && ss -tlnp | grep -E ":(80|443|3000|8080|9925)" && \
echo -e "\n=== DNS ===" && dig al.fluidgrid.site +short && \
echo -e "\n=== Local Test ===" && curl -s http://localhost:3000 | head -5

# Watch logs in real-time while testing
tail -f /var/log/caddy/al.access.log & \
journalctl -u caddy -f
```

---

## Next Steps After Fixing

Once the site is working:
1. Disable debug mode by removing `debug` from global options
2. Change log level from `DEBUG` to `INFO` or `WARN`
3. Remove or comment out diagnostic endpoints
4. Keep the debug headers if helpful, or remove them

To disable debugging, edit Caddyfile:
```caddy
{
    # debug  # <-- Comment this out
    admin localhost:2019  # Change to localhost for security
}
```

And change log levels:
```caddy
log {
    level INFO  # Change from DEBUG to INFO
}
```

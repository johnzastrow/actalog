# Fix: Caddy Using Wrong Configuration File

## Problem Identified

Caddy is using `/etc/caddy/Caddyfile` (the default config), but your ActaLog configuration is in `/home/jcz/Github/actionlog/Caddyfile`.

**Current situation:**
- ✗ Caddy is reading: `/etc/caddy/Caddyfile` (default config - just serves static files on port 80)
- ✓ Your config is at: `/home/jcz/Github/actionlog/Caddyfile` (with al.fluidgrid.site setup + debugging)

## Solution: Replace Caddy's Config

Run these commands to replace Caddy's config with your enhanced version:

```bash
# 1. Backup the current Caddy config
sudo cp /etc/caddy/Caddyfile /etc/caddy/Caddyfile.backup.$(date +%Y%m%d)

# 2. Copy your enhanced config to Caddy's config location
sudo cp /home/jcz/Github/actionlog/Caddyfile /etc/caddy/Caddyfile

# 3. Verify the config is valid
sudo caddy validate --config /etc/caddy/Caddyfile

# 4. Reload Caddy to use the new config
sudo systemctl reload caddy

# 5. Check Caddy status
sudo systemctl status caddy
```

## Verify It's Working

After reloading, check:

```bash
# 1. Check Caddy logs for any errors
sudo journalctl -u caddy -n 50 --no-pager

# 2. Test the diagnostic endpoint
curl -v https://al.fluidgrid.site/caddy-test

# 3. Check debug headers
curl -I https://al.fluidgrid.site/

# 4. Watch logs in real-time
sudo journalctl -u caddy -f
```

## What's in the Enhanced Config

Your enhanced Caddyfile includes:
- ✅ Configuration for `al.fluidgrid.site` → proxies to `localhost:3000`
- ✅ Configuration for `recipe.fluidgrid.site` → proxies to `localhost:9925`
- ✅ Global debug mode enabled
- ✅ Admin API on port 2019
- ✅ DEBUG-level logging
- ✅ Diagnostic endpoints at `/caddy-test`
- ✅ Debug response headers
- ✅ HTTP port 80 catch-all for troubleshooting

## Alternative: Symlink Approach

If you want to keep your Caddyfile in the project directory and have Caddy use it:

```bash
# 1. Backup current config
sudo cp /etc/caddy/Caddyfile /etc/caddy/Caddyfile.backup.$(date +%Y%m%d)

# 2. Remove the current config
sudo rm /etc/caddy/Caddyfile

# 3. Create a symlink
sudo ln -s /home/jcz/Github/actionlog/Caddyfile /etc/caddy/Caddyfile

# 4. Reload Caddy
sudo systemctl reload caddy
```

**Pros:** Edit one file, Caddy automatically uses it
**Cons:** If you move the project directory, the symlink breaks

## Quick One-Liner

```bash
sudo cp /etc/caddy/Caddyfile /etc/caddy/Caddyfile.backup.$(date +%Y%m%d) && \
sudo cp /home/jcz/Github/actionlog/Caddyfile /etc/caddy/Caddyfile && \
sudo caddy validate --config /etc/caddy/Caddyfile && \
sudo systemctl reload caddy && \
echo "✓ Caddy config updated and reloaded" && \
sudo systemctl status caddy --no-pager
```

## Expected Output After Fix

### Before (default config):
```bash
$ curl http://your-server-ip
# Shows default Caddy page or directory listing
```

### After (your config):
```bash
$ curl http://your-server-ip
Caddy is running. Requested host: your-server-ip - Try https://al.fluidgrid.site or https://recipe.fluidgrid.site

$ curl https://al.fluidgrid.site/caddy-test
Caddy is serving al.fluidgrid.site - 2025-11-20 19:10:45
```

## Troubleshooting

### "caddy validate" shows errors
Check the error message. Common issues:
- Syntax errors in Caddyfile
- Port already in use
- Invalid directives

### Caddy fails to reload
```bash
# Check what went wrong
sudo journalctl -u caddy -n 100 --no-pager | grep -i error

# Try a full restart instead
sudo systemctl restart caddy
```

### "Permission denied" errors
The Caddyfile needs to be readable by the Caddy user:
```bash
sudo chmod 644 /etc/caddy/Caddyfile
sudo chown root:root /etc/caddy/Caddyfile
```

## Keeping Configs in Sync

If you edit the project Caddyfile, remember to copy it to /etc/caddy/:
```bash
# After editing /home/jcz/Github/actionlog/Caddyfile
sudo cp /home/jcz/Github/actionlog/Caddyfile /etc/caddy/Caddyfile
sudo systemctl reload caddy
```

Or create a helper script:
```bash
cat > ~/update-caddy-config.sh << 'EOF'
#!/bin/bash
sudo cp /home/jcz/Github/actionlog/Caddyfile /etc/caddy/Caddyfile
sudo caddy validate --config /etc/caddy/Caddyfile && \
sudo systemctl reload caddy && \
echo "✓ Caddy config updated successfully"
EOF
chmod +x ~/update-caddy-config.sh
```

# Port Conflict Handling in start-frontend.sh

## Overview

The `scripts/start-frontend.sh` script now automatically detects and handles port conflicts on port 3000.

## What Was Added

### 1. **Port Detection**
The script now checks if port 3000 is already in use before starting the frontend. It uses multiple methods for compatibility:
- `lsof` (preferred - most detailed information)
- `ss` (fallback)
- `netstat` (second fallback)

### 2. **Process Information Display**
When a port conflict is detected, the script shows:
```
⚠ WARNING: Port 3000 is already in use!
  Process ID:   12345
  Process name: node
  Command:      npm run dev
```

### 3. **Interactive Options**
Three options are presented:

#### Option 1: Stop Existing Process and Use Port 3000
- Automatically kills the process using port 3000
- Falls back to `sudo` if regular kill fails
- Waits 2 seconds for port to be released
- Proceeds with starting ActaLog on port 3000

#### Option 2: Start ActaLog on Different Port
- Scans ports 3001-3010 to find an available port
- Automatically uses the first available port
- Displays the port number clearly
- Shows updated access URL
- Warns to update reverse proxy configuration if applicable

#### Option 3: Cancel and Exit
- Safe exit without making any changes

### 4. **Port Number Display**
A clear summary is shown before starting:
```
=========================================
Starting ActaLog Frontend
=========================================
 Port:     3001
 Mode:     Development
 Hostname: localhost

⚠ Using non-standard port: 3001
   Access URL: http://localhost:3001
=========================================
```

### 5. **Reverse Proxy Warning**
If using a non-standard port with a reverse proxy (Caddy), the script reminds you:
```
⚠ Remember to update your reverse proxy configuration!
   Change 'reverse_proxy localhost:3000' to 'reverse_proxy localhost:3001'
   in your Caddyfile or nginx config.
```

## Usage Examples

### Scenario 1: Port 3000 is Free
```bash
$ ./scripts/start-frontend.sh

ActaLog frontend starter — guided mode
Run in (d)ev or (p)review (production) mode? [d/p]: d
...
Summary:
 Mode:       dev
 Hostname:   localhost
 HTTPS flag: none

=========================================
Starting ActaLog Frontend
=========================================
 Port:     3000
 Mode:     Development
 Hostname: localhost
=========================================

Running: npm run dev -- --port 3000
```

### Scenario 2: Port 3000 is in Use - Kill Existing Process
```bash
$ ./scripts/start-frontend.sh

ActaLog frontend starter — guided mode
Run in (d)ev or (p)review (production) mode? [d/p]: d
...
Summary:
 Mode:       dev
 Hostname:   localhost
 HTTPS flag: none

⚠ WARNING: Port 3000 is already in use!
  Process ID:   54321
  Process name: node
  Command:      /usr/bin/node /home/user/other-project/server.js

Options:
  1) Stop the existing process and start ActaLog on port 3000
  2) Start ActaLog on a different port
  3) Cancel and exit

Choose an option [1/2/3]: 1

Attempting to stop process on port 3000...
Killing process 54321...
✓ Process stopped. Proceeding with port 3000...

=========================================
Starting ActaLog Frontend
=========================================
 Port:     3000
 Mode:     Development
 Hostname: localhost
=========================================

Running: npm run dev -- --port 3000
```

### Scenario 3: Port 3000 is in Use - Use Different Port
```bash
$ ./scripts/start-frontend.sh

ActaLog frontend starter — guided mode
Run in (d)ev or (p)review (production) mode? [d/p]: d
Will you access the frontend on this machine's localhost or a domain name? [l=localhost / D=domain]: D
Enter the domain you'll use (e.g. actalog.example.com): al.fluidgrid.site
...
Will you use a reverse proxy like Caddy to handle HTTPS? [Y/n]: Y

Summary:
 Mode:       dev
 Hostname:   al.fluidgrid.site
 HTTPS flag: none

⚠ WARNING: Port 3000 is already in use!
  Process ID:   54321
  Process name: node
  Command:      npm run dev

Options:
  1) Stop the existing process and start ActaLog on port 3000
  2) Start ActaLog on a different port
  3) Cancel and exit

Choose an option [1/2/3]: 2

Finding an available port...
✓ Found available port: 3001

📌 IMPORTANT: ActaLog will start on port 3001
   Access it at: http://al.fluidgrid.site:3001

   ⚠ Remember to update your reverse proxy configuration!
   Change 'reverse_proxy localhost:3000' to 'reverse_proxy localhost:3001'
   in your Caddyfile or nginx config.

=========================================
Starting ActaLog Frontend
=========================================
 Port:     3001
 Mode:     Development
 Hostname: al.fluidgrid.site

⚠ Using non-standard port: 3001
   Access URL: http://al.fluidgrid.site:3001
=========================================

Running: npm run dev -- --host --port 3001
```

## Manual Port Testing

You can manually check if a port is in use:

```bash
# Check if port 3000 is in use
lsof -ti:3000

# Get detailed info about what's using the port
lsof -i:3000

# Alternative with ss
ss -tlnp | grep :3000

# Alternative with netstat
netstat -tlnp | grep :3000
```

## Updating Reverse Proxy for Different Ports

If you choose option 2 and ActaLog starts on a different port, update your reverse proxy:

### Caddy (Caddyfile)
```caddy
al.fluidgrid.site {
    # Change from:
    # reverse_proxy localhost:3000

    # To (using new port):
    reverse_proxy localhost:3001

    # ... rest of config
}
```

Then reload Caddy:
```bash
sudo systemctl reload caddy
# Or:
caddy reload
```

### Nginx
```nginx
location / {
    # Change from:
    # proxy_pass http://localhost:3000;

    # To (using new port):
    proxy_pass http://localhost:3001;
}
```

Then reload Nginx:
```bash
sudo systemctl reload nginx
# Or:
sudo nginx -s reload
```

## Troubleshooting

### "Cannot automatically kill process (lsof not available)"
Install lsof:
```bash
# Ubuntu/Debian
sudo apt-get install lsof

# CentOS/RHEL
sudo yum install lsof

# macOS (usually pre-installed)
brew install lsof
```

### "Failed to kill process. You may need sudo privileges"
Some processes require elevated privileges to terminate:
- System services
- Processes owned by other users
- Processes started with sudo

The script will prompt you to use sudo automatically.

### "Could not find an available port in range 3001-3010"
All ports in the range are occupied. You can:
1. Manually stop some services
2. Modify the script to check a wider range
3. Manually specify a port when running npm:
   ```bash
   cd web
   npm run dev -- --port 4000
   ```

## Best Practices

1. **Always use the script** - It handles port conflicts gracefully
2. **Update reverse proxy** - Remember to update Caddy/Nginx if using a different port
3. **Document port changes** - Keep track of which port you're using
4. **Prefer port 3000** - Use option 1 (kill existing) if the conflicting process isn't needed
5. **Clean up** - Stop unused development servers to free up ports

## Files Modified

- ✅ `scripts/start-frontend.sh` - Added port conflict detection and handling
- ✅ Made executable with `chmod +x`
- ✅ Syntax validated with `bash -n`

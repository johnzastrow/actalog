# start-frontend.sh Version History

## Version 1.2.0 (2025-11-20)

**Added:**
- Version number displayed when script runs
- Version shown in header: "ActaLog Frontend Starter v1.2.0"
- Version shown in final configuration summary
- Environment variable integration with vite.config.js:
  - Exports `VITE_DEV_HOST`, `VITE_DEV_PORT`, `VITE_USE_HTTPS`, `VITE_DEPLOYMENT_URL`
  - Proper deployment URL construction based on mode (localhost/domain/proxy)

**Fixed:**
- Echo statements with `\n` now properly use separate echo commands
- Better output formatting with proper newlines
- Clearer configuration summary display

**Purpose:**
- Version tracking helps identify which script version is running on remote servers
- Ensures latest changes have been pulled to deployment environments

## Version 1.1.0 (2025-11-20)

**Added:**
- Port conflict detection and handling
- Interactive options when port 3000 is in use:
  1. Kill existing process and use port 3000
  2. Find alternative port (scans 3001-3010)
  3. Cancel and exit
- Process information display (PID, name, command)
- Automatic port fallback with availability scanning
- Reverse proxy configuration warnings when using non-standard ports
- Support for `lsof`, `ss`, and `netstat` for port detection

**Documentation:**
- Created `PORT_CONFLICT_HANDLING.md`

## Version 1.0.0 (2025-11-19)

**Initial Features:**
- Guided mode for configuring frontend startup
- Development vs Preview mode selection
- Localhost vs Domain configuration
- LAN exposure option for localhost mode
- HTTPS support with mkcert integration
- Domain mode with DNS verification
- Reverse proxy mode detection
- Automatic cert generation with mkcert
- HMR (Hot Module Reload) configuration
- Interactive prompts for all configuration options

**Documentation:**
- Initial script documentation in comments
- Usage examples in script header

---

## Version Numbering

The script follows semantic versioning:
- **Major (X.0.0)**: Breaking changes or major rewrites
- **Minor (1.X.0)**: New features, enhancements
- **Patch (1.0.X)**: Bug fixes, minor improvements

## How to Use Version Information

### Check Current Version
When you run the script, the version is displayed at the top:
```
ActaLog Frontend Starter v1.2.0
=========================================
```

And again in the final summary:
```
=========================================
Starting ActaLog Frontend v1.2.0
=========================================
```

### Track Deployment Versions
When deploying to a remote server:
1. Note the version number from script output
2. Compare with version in git repository
3. If versions don't match, pull latest changes:
   ```bash
   cd /path/to/actionlog
   git pull origin main
   ./scripts/start-frontend.sh
   ```

### Verify Script Updates
After pulling changes from git:
```bash
# Check script version
head -20 scripts/start-frontend.sh | grep SCRIPT_VERSION

# Run script to see version in action
./scripts/start-frontend.sh
```

## Updating the Version

When making changes to the script, increment the version:

1. **For bug fixes:**
   ```bash
   # Change from 1.2.0 to 1.2.1
   sed -i 's/SCRIPT_VERSION="1.2.0"/SCRIPT_VERSION="1.2.1"/' scripts/start-frontend.sh
   ```

2. **For new features:**
   ```bash
   # Change from 1.2.0 to 1.3.0
   sed -i 's/SCRIPT_VERSION="1.2.0"/SCRIPT_VERSION="1.3.0"/' scripts/start-frontend.sh
   ```

3. **For breaking changes:**
   ```bash
   # Change from 1.2.0 to 2.0.0
   sed -i 's/SCRIPT_VERSION="1.2.0"/SCRIPT_VERSION="2.0.0"/' scripts/start-frontend.sh
   ```

4. **Update this history file** with change notes

## Integration with Git

### Commit Messages
When incrementing version, include version in commit:
```bash
git add scripts/start-frontend.sh scripts/START_FRONTEND_VERSION_HISTORY.md
git commit -m "feat: Bump start-frontend.sh to v1.3.0 - Add feature X"
```

### Git Tags (Optional)
You can also tag script versions:
```bash
git tag -a script-frontend-v1.2.0 -m "Frontend starter script v1.2.0"
git push origin script-frontend-v1.2.0
```

## Related Files

- `scripts/start-frontend.sh` - The main script
- `VITE_CONFIG_INTEGRATION.md` - How script integrates with Vite config
- `PORT_CONFLICT_HANDLING.md` - Port conflict resolution documentation
- `CADDY_CONFIG_FIX.md` - Caddy reverse proxy configuration
- `CADDY_DEBUGGING.md` - Caddy debugging guide

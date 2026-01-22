# Class Scheduling Feature

ActaLog includes an optional class scheduling system for gyms that offer scheduled classes, personal training sessions, or group workouts.

## Enabling Scheduling

Set the `SCHEDULER_ENABLED` environment variable to `true`:

### Docker Run

```bash
docker run -e SCHEDULER_ENABLED=true \
  -e DB_DRIVER=postgres \
  -e DB_HOST=localhost \
  ... \
  ghcr.io/johnzastrow/actalog:dev
```

### Docker Compose

In your `docker/.env` file:

```env
SCHEDULER_ENABLED=true
```

### Local Development

In your `.env` file (project root):

```env
SCHEDULER_ENABLED=true
```

## Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `SCHEDULER_ENABLED` | `false` | Enable/disable the scheduling feature |
| `SCHEDULER_INTERVAL` | `6h` | How often the session materializer runs |
| `SCHEDULER_DAYS_AHEAD` | `14` | Number of days ahead to pre-create sessions |
| `SCHEDULER_RUN_ON_STARTUP` | `true` | Run session materialization immediately on startup |

### Example Configuration

```env
# Enable scheduling
SCHEDULER_ENABLED=true

# Create sessions 3 weeks ahead instead of default 2
SCHEDULER_DAYS_AHEAD=21

# Run materializer every 12 hours instead of 6
SCHEDULER_INTERVAL=12h

# Don't run on startup (wait for first scheduled interval)
SCHEDULER_RUN_ON_STARTUP=false
```

## What Changes When Scheduling is Enabled

### User Interface

1. **Navigation Bar**: The center button changes from a `+` (quick-log) to a calendar icon that navigates to `/schedule`

2. **Dashboard**: The "Log Workout" button is removed (users use the Quick Log button instead)

3. **Admin Profile**: Shows "Scheduling enabled. Disable in the server." in green text under the Administration section

4. **Admin Menu**: Class Scheduling and Class Packages menu items become functional

### Backend

1. **Session Materializer**: A background job runs periodically to create individual session instances from recurring schedule templates

2. **API Endpoints**: Scheduling-related endpoints become active:
   - `/api/schedule/*` - User schedule views
   - `/api/admin/scheduling/*` - Admin schedule management
   - `/api/coach/*` - Coach dashboard and session management

## How Scheduling Works

### Concepts

- **Schedule Template**: A recurring class definition (e.g., "CrossFit WOD" every Monday/Wednesday/Friday at 6 AM)
- **Session**: A single instance of a class on a specific date/time, created from a template
- **Materialization**: The process of creating individual sessions from templates

### Session Materializer

The materializer runs on a schedule (default: every 6 hours) and:

1. Looks at all active schedule templates
2. Creates individual session records for the configured number of days ahead
3. Skips sessions that already exist (idempotent)
4. Handles recurring patterns (daily, weekly, specific days)

### Example Flow

1. Admin creates a schedule template: "Morning CrossFit" on Mon/Wed/Fri at 6:00 AM
2. Materializer runs and creates sessions for the next 14 days
3. Users see available sessions in the schedule view
4. Users can book/reserve spots in sessions
5. Coaches see their assigned sessions in the coach dashboard

## Admin Setup

### 1. Create Locations

Before creating schedules, define your gym locations:

- Navigate to Admin → Class Scheduling → Locations
- Add locations (e.g., "Main Floor", "Weight Room", "Studio A")

### 2. Assign Coaches

Coaches are users with coach assignments:

- Navigate to Admin → Class Scheduling → Coaches
- Assign users as coaches for specific organizations

### 3. Create Schedule Templates

Define your recurring class schedule:

- Navigate to Admin → Class Scheduling → Templates
- Create templates with:
  - Name (e.g., "Morning CrossFit")
  - Location
  - Day(s) of week
  - Start time and duration
  - Capacity (max participants)
  - Assigned coach(es)

### 4. Manage Sessions

Once templates are created, sessions are auto-generated:

- Navigate to Admin → Class Scheduling → Sessions
- View, edit, or cancel individual sessions
- Handle one-off changes without affecting the template

## Class Packages (Optional)

If using a credit-based booking system:

- Navigate to Admin → Class Packages
- Create packages (e.g., "10 Class Pack", "Unlimited Monthly")
- Users purchase packages to get credits
- Credits are consumed when booking sessions

## Verifying Scheduling Status

### Check via API

```bash
curl http://localhost:8080/api/version | jq
```

Response includes:
```json
{
  "version": "1.1.0-beta",
  "scheduling_enabled": true
}
```

### Check in UI

1. Look at the center navigation button:
   - Calendar icon = scheduling enabled
   - Plus icon = scheduling disabled

2. Go to Profile → Administration section:
   - Green text: "Scheduling enabled. Disable in the server."
   - Gray text: "Scheduling disabled. Enable in the server."

## Disabling Scheduling

To disable scheduling, set:

```env
SCHEDULER_ENABLED=false
```

Or simply remove the `SCHEDULER_ENABLED` variable (defaults to `false`).

**Note**: Disabling scheduling does not delete existing schedule data. It only:
- Changes the UI back to quick-log mode
- Stops the session materializer from running
- Hides scheduling-related admin features

## Troubleshooting

### Sessions Not Being Created

1. Check that `SCHEDULER_ENABLED=true` is set
2. Verify the materializer is running (check logs for "session materialization" messages)
3. Ensure schedule templates exist and are active
4. Check `SCHEDULER_DAYS_AHEAD` - sessions are only created this many days in advance

### Materializer Running Too Often/Not Often Enough

Adjust `SCHEDULER_INTERVAL`:

```env
# Run every hour
SCHEDULER_INTERVAL=1h

# Run every 24 hours
SCHEDULER_INTERVAL=24h
```

### Want Sessions Created Further in Advance

Increase `SCHEDULER_DAYS_AHEAD`:

```env
# Create sessions 30 days ahead
SCHEDULER_DAYS_AHEAD=30
```

## Related Documentation

- [DEPLOYMENT.md](./DEPLOYMENT.md) - General deployment guide
- [DOCKER.md](./DOCKER.md) - Docker-specific configuration
- [DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md) - Scheduling-related tables
- [USER_PERMISSIONS.md](./USER_PERMISSIONS.md) - Coach and admin permissions

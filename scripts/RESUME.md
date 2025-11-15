# Development Resume - v0.4.0-beta

**Date Paused:** 2025-11-12
**Version:** 0.4.0-beta (partial completion)
**Status:** Backend complete with seeded data. Frontend partially implemented (stores + WOD Library).

## Quick Start to Resume Development

```bash
# Start backend (from project root)
make run
# Backend will be available at http://localhost:8080

# Start frontend (in separate terminal)
cd web
npm run dev
# Frontend will be available at http://localhost:3000
```

## What's Complete in v0.4.0-beta

### Backend (100% Complete)

**Database & Migrations:**
- ✅ Migration v0.4.0 applied: `wods` table, `workout_wods` junction table
- ✅ Multi-database support (SQLite, PostgreSQL, MySQL)
- ✅ Idempotent seeding with sentinel checks

**Seeded Data:**
- ✅ 31 movements (Back Squat, Deadlift, Pull-ups, etc.)
- ✅ 10 WODs:
  - 8 Girl WODs: Fran, Helen, Cindy, Grace, Annie, Karen, Diane, Elizabeth
  - 2 Hero WODs: Murph, DT
- ✅ 3 workout templates:
  - Strength Training - Back Squat Focus (1 movement)
  - Olympic Lifting - Clean & Jerk Practice (2 movements)
  - Gymnastics Strength (0 movements - movements failed to seed)

**API Endpoints Working:**
```
GET    /api/wods              - List all WODs
GET    /api/wods/{id}         - Get WOD by ID
GET    /api/wods/search?q=    - Search WODs by name
POST   /api/wods              - Create custom WOD
PUT    /api/wods/{id}         - Update custom WOD
DELETE /api/wods/{id}         - Delete custom WOD

GET    /api/templates         - List all templates
GET    /api/templates/{id}    - Get template by ID
POST   /api/templates         - Create template
PUT    /api/templates/{id}    - Update template
DELETE /api/templates/{id}    - Delete template
GET    /api/workouts/my-templates - Get user's custom templates

GET    /api/movements         - List movements
POST   /api/movements         - Create custom movement
```

**Code Files (Backend):**
- `internal/domain/wod.go` - WOD entity and interface
- `internal/domain/workout_wod.go` - Workout-WOD junction model
- `internal/repository/wod_repository.go` - WOD data access (197 lines)
- `internal/repository/workout_wod_repository.go` - Junction table ops
- `internal/service/wod_service.go` - WOD business logic (360 lines)
- `internal/service/workout_wod_service.go` - Linking logic (192 lines)
- `internal/handler/wod_handler.go` - WOD HTTP handlers (247 lines)
- `internal/handler/workout_wod_handler.go` - Linking handlers (219 lines)
- `internal/repository/database.go` - Seeding functions (lines 637-934)

### Frontend (40% Complete)

**Pinia Stores Created:**
- ✅ `web/src/stores/wods.js` (155 lines)
  - Actions: fetchWods(), fetchWodById(), searchWods(), createWod(), updateWod(), deleteWod()
  - Filters: filterByType(), filterBySource(), getStandardWods(), getCustomWods()
- ✅ `web/src/stores/templates.js` (227 lines)
  - Actions: fetchTemplates(), fetchTemplateById(), fetchMyTemplates(), createTemplate(), updateTemplate(), deleteTemplate()
  - WOD linking: fetchTemplateWods(), addWodToTemplate(), removeWodFromTemplate(), toggleWodPR()
  - Filters: getStandardTemplates(), getCustomTemplates(), getTemplatesWithMovementCount()

**Views Updated:**
- ✅ `web/src/views/WODLibraryView.vue` - Updated to use `useWodsStore` (295 lines)
  - Browse WODs with type filtering
  - Search functionality
  - Create/edit custom WODs

## What's NOT Complete

### Frontend Work Remaining (Next Priority)

1. **Dashboard Integration** (web/src/views/DashboardView.vue)
   - Add WOD count card
   - Add template count card
   - Show recent templates in addition to recent workouts
   - Link to WOD Library and Template Library

2. **Template Library View** (NEW FILE NEEDED)
   - Create `web/src/views/TemplateLibraryView.vue`
   - Browse all standard and custom templates
   - Filter by type/source
   - View template details (movements + WODs)
   - Create/edit custom templates
   - Link WODs to templates

3. **Workout Logging with Templates** (web/src/views/LogWorkoutView.vue)
   - Add "Use Template" button
   - Template selection dialog
   - Pre-populate form with template's movements and WODs
   - Allow customization before saving
   - Link logged workout to template (if used)

4. **Router Configuration** (web/src/router/index.js)
   - Add route for Template Library view (if not already present)
   - Ensure proper navigation guards

### Testing & Validation

Currently **NOT tested** (critical for resuming):
- [ ] Test WOD Library view in browser (http://localhost:3000/wods)
- [ ] Verify WOD creation/editing works
- [ ] Test template fetching from API
- [ ] Verify navigation between views
- [ ] Check mobile responsiveness

## File Locations Reference

### Backend Files
```
internal/
├── domain/
│   ├── wod.go                         # WOD entity + interface
│   ├── workout_wod.go                 # Junction model
│   └── workout.go                     # Template model
├── repository/
│   ├── database.go                    # Seeding (lines 637-934)
│   ├── wod_repository.go              # WOD CRUD
│   ├── workout_repository.go          # Template CRUD
│   ├── workout_wod_repository.go      # Linking operations
│   └── movement_repository.go         # Movement ops
├── service/
│   ├── wod_service.go                 # WOD logic (360 lines)
│   ├── workout_service.go             # Template logic (382 lines)
│   └── workout_wod_service.go         # Linking logic (192 lines)
└── handler/
    ├── wod_handler.go                 # WOD endpoints (247 lines)
    ├── workout_template_handler.go    # Template endpoints
    └── workout_wod_handler.go         # Linking endpoints (219 lines)
```

### Frontend Files
```
web/src/
├── stores/
│   ├── wods.js                        # WOD state (155 lines) ✅
│   ├── templates.js                   # Template state (227 lines) ✅
│   └── auth.js                        # Auth state (existing)
├── views/
│   ├── WODLibraryView.vue             # Browse WODs ✅ UPDATED
│   ├── DashboardView.vue              # Dashboard ⏳ NEEDS UPDATE
│   ├── LogWorkoutView.vue             # Log workouts ⏳ NEEDS UPDATE
│   └── TemplateLibraryView.vue        # ❌ NEEDS CREATION
└── router/
    └── index.js                       # Routes ⏳ MAY NEED UPDATE
```

## Key Technical Decisions

1. **Template vs Instance Architecture:**
   - Workouts are now **templates** (reusable plans)
   - User workouts are **instances** of templates with actual performance data
   - This enables saving favorite workouts and tracking performance over time

2. **WOD Integration:**
   - WODs can be standalone or linked to workout templates
   - Many-to-many relationship via `workout_wods` junction table
   - Each WOD in a template can have PR tracking flag

3. **State Management:**
   - Pinia stores for all frontend state
   - Composition API pattern (`ref`, `computed`)
   - Proper error handling with user-friendly messages
   - Loading states for UI feedback

4. **Seeding Strategy:**
   - Idempotent seeding checks for existing data before inserting
   - Uses sentinel records (e.g., first WOD "Fran") to determine if seeding ran
   - Separate functions for movements, WODs, and templates

## Common Commands

```bash
# Backend
make build          # Build binary
make run            # Run backend server (:8080)
make test           # Run tests
make lint           # Run linter

# Frontend
npm run dev         # Start dev server (:3000)
npm run build       # Production build
npm run lint        # Run ESLint
npm run format      # Format with Prettier

# Database
sqlite3 actalog.db  # Open SQLite database
.tables             # List all tables
.schema wods        # Show WOD table schema
```

## Testing Checklist for Next Session

1. **Verify Backend:**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/api/wods | jq
   curl http://localhost:8080/api/templates | jq
   ```

2. **Test WOD Library:**
   - Navigate to http://localhost:3000/wods
   - Verify 10 WODs display
   - Test filtering by type (Girl, Hero)
   - Test search functionality
   - Verify clicking on WOD shows details

3. **Check Console:**
   - Open browser DevTools console
   - Look for any errors when navigating views
   - Verify API calls succeed (Network tab)

## Next Steps (Priority Order)

1. **Test current implementation** (30 min)
   - Launch both servers
   - Navigate to /wods route
   - Verify data displays correctly
   - Check browser console for errors

2. **Create Template Library view** (2 hours)
   - Copy pattern from WODLibraryView.vue
   - Use `useTemplatesStore`
   - Display templates with movement counts
   - Add filtering and search

3. **Update Dashboard** (1 hour)
   - Add template/WOD count cards
   - Link to new Library views
   - Show recent templates

4. **Update LogWorkoutView** (3 hours)
   - Add "Use Template" button
   - Template selection dialog
   - Pre-populate form from template
   - Handle both movements and WODs

5. **End-to-end testing** (1 hour)
   - Test full workflow: browse → select template → log workout
   - Verify data persistence
   - Check mobile responsiveness

## Known Issues

1. **Gymnastics Strength template** has 0 movements
   - Seeding failed for this template's movements
   - Backend seeds 3 templates but only 2 have working movements
   - Need to investigate why movement linking failed

2. **Frontend dev server** (`npm run dev`)
   - May show Vite warnings on first load
   - Hot reload works correctly
   - No blocking issues

3. **Package.json warning**
   - Modified during version bump
   - If running `npm install` shows warnings, they're non-critical

## Environment Variables

Required in `.env` file:
```env
DB_DRIVER=sqlite3
DB_NAME=actalog.db
JWT_SECRET=your-secret-key
```

## Git Status

Current branch: `main`

Changes staged:
- ✅ pkg/version/version.go (v0.4.0-beta)
- ✅ web/package.json (v0.4.0)
- ✅ CHANGELOG.md (v0.4.0 entry added)
- ✅ docs/TODO.md (status updated)

Modified but not staged:
- web/package-lock.json (auto-generated during version bump)

New files:
- web/src/stores/wods.js
- web/src/stores/templates.js
- RESUME.md (this file)

**Recommendation:** Commit current state before resuming:
```bash
git add .
git commit -m "feat(v0.4.0): Add WOD system, templates, and frontend stores

- Added WOD management with 10 seeded standard WODs
- Created workout template system with WOD linking
- Implemented Pinia stores for WODs and templates
- Updated WOD Library view to use new store architecture
- Seeded 3 workout templates with movements
- Version bumped to 0.4.0-beta

In Progress:
- Dashboard integration
- Template Library view
- Template-based workout logging"
```

## Documentation

Key docs to reference:
- `CLAUDE.md` - Development guidelines and commands
- `docs/DATABASE_SCHEMA.md` - Complete database schema
- `docs/ARCHITECTURE.md` - System architecture
- `CHANGELOG.md` - Version history
- `docs/TODO.md` - Current tasks and priorities

## Contact Points for Questions

When resuming, check:
1. Is backend running? (`curl http://localhost:8080/health`)
2. Is frontend dev server running? (http://localhost:3000)
3. Are there any console errors in browser?
4. Did package-lock.json change cause issues? (`npm install` to fix)
5. Do tests pass? (`make test`)

---

**Ready to resume:** Follow "Quick Start" section → Complete "Testing Checklist" → Continue with "Next Steps" priority #2 (Create Template Library view).

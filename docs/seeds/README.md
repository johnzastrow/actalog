# Seed Data CSV Files

This directory contains CSV seed files for ActaLog standard data that can be used for:
1. Testing import/export functionality
2. Loading standard WODs and Movements in new deployments
3. Providing reference CSV format examples

## Files

### movements.csv
Standard CrossFit movements including:
- **Weightlifting**: Back Squat, Deadlift, Clean, Snatch, etc. (17 movements)
- **Gymnastics**: Pull-ups, Muscle-ups, Handstand Push-ups, etc. (7 movements)
- **Bodyweight**: Push-ups, Sit-ups, Burpees, etc. (5 movements)
- **Cardio**: Row, Run, Bike, Ski Erg (4 movements)

**Total**: 33 standard movements

**CSV Format**:
```csv
name,type,description,is_standard,created_by_email
```

**Field Descriptions**:
- `name` (required): Movement name
- `type` (required): One of: weightlifting, cardio, gymnastics, bodyweight
- `description` (optional): Movement description
- `is_standard` (required): true for standard movements, false for custom
- `created_by_email` (optional): Email of user who created (empty for standard)

### wods.csv
Famous CrossFit benchmark WODs including:
- **Girls**: Fran, Helen, Cindy, Grace, Annie, Karen, Diane, Elizabeth
- **Heroes**: Murph, DT

**Total**: 10 standard WODs

**CSV Format**:
```csv
name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
```

**Field Descriptions**:
- `name` (required): WOD name
- `source` (required): One of: CrossFit, Other Coach, Self-recorded
- `type` (required): One of: Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created
- `regime` (required): One of: EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills
- `score_type` (required): One of: Time (HH:MM:SS), Rounds+Reps, Max Weight
- `description` (required): WOD description
- `url` (optional): Reference URL
- `notes` (optional): Additional notes
- `is_standard` (required): true for standard WODs, false for custom
- `created_by_email` (optional): Email of user who created (empty for standard)

## Usage

### Import via API (Future)
```bash
# Import movements
curl -X POST http://localhost:8080/api/import/movements/preview \
  -F "file=@seeds/movements.csv" \
  -H "Authorization: Bearer <token>"

# Import WODs
curl -X POST http://localhost:8080/api/import/wods/preview \
  -F "file=@seeds/wods.csv" \
  -H "Authorization: Bearer <token>"
```

### Export via API (Future)
```bash
# Export movements
curl http://localhost:8080/api/export/movements?format=csv \
  -H "Authorization: Bearer <token>" \
  > movements_export.csv

# Export WODs
curl http://localhost:8080/api/export/wods?format=csv \
  -H "Authorization: Bearer <token>" \
  > wods_export.csv
```

## CSV Format Rules

1. **Headers**: First row must contain column names (case-sensitive)
2. **Quoting**: Fields with commas must be quoted (e.g., `"21-15-9 reps for time of: Thrusters, Pull-ups"`)
3. **Empty Fields**: Leave empty for optional fields (but include the comma)
4. **Boolean Values**: Use `true` or `false` (lowercase)
5. **Encoding**: UTF-8 encoding required
6. **Line Endings**: Unix (LF) or Windows (CRLF) both supported

## Testing Round-Trip

To verify import/export round-trip functionality:

1. Export data to CSV
2. Import the exported CSV
3. Compare exported CSV with original - should match exactly

## Adding Custom Data

To add custom movements or WODs:
- Set `is_standard` to `false`
- Set `created_by_email` to the user's email address
- Admin users can import as standard by setting `is_standard` to `true`

## Validation Rules

**Movements**:
- `name` must be unique
- `type` must be one of the valid enum values
- `created_by_email` must exist in users table (if provided)

**WODs**:
- `name` must be unique
- All enum fields must match valid values
- `created_by_email` must exist in users table (if provided)

## Notes

- These CSV files match the format used by the import/export API endpoints
- Standard movements/WODs are also seeded programmatically via `internal/repository/database.go`
- The CSV format is designed for human readability and spreadsheet compatibility
- For complete data exports including user workouts, use JSON format instead

**Version**: 0.5.1-beta
**Last Updated**: 2025-11-21

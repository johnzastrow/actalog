# ActaLog Seed Data Files

This directory contains CSV seed files for populating ActaLog with standard CrossFit movements and benchmark WODs.

## Files

### movements.csv
Contains 95 standard CrossFit movements including:
- **Olympic Lifts**: Snatch, Clean, Jerk, Clean & Jerk (and variations)
- **Weightlifting**: Squats, Deadlifts, Presses, Thrusters, Romanian Deadlift, Hip Thrust
- **Gymnastics**: Pull-ups, Muscle-ups, Handstand Push-ups, Rope Climbs, Handstand Walk, Wall Walk, L-Sit, Ring Row
- **Bodyweight**: Push-ups, Squats, Burpees, Box Jumps, Hollow Hold, Arch Hold, Broad Jump
- **Cardio**: Row, Run, Bike, Ski Erg, Jump Rope, Battle Ropes, Airdyne Bike
- **Strongman**: Sandbag Carry, Yoke Carry, Atlas Stone Lift, Tire Flip

**CSV Structure:**
```
id,name,description,type,is_standard,created_by
```

**Field Descriptions:**
- `id`: Unique identifier (integer)
- `name`: Movement name (string)
- `description`: Detailed description of the movement (string)
- `type`: Movement category - `weightlifting`, `gymnastics`, `bodyweight`, or `cardio`
- `is_standard`: Always `TRUE` for standard movements
- `created_by`: NULL for standard movements (user ID for custom movements)

### wods.csv
Contains 68 famous CrossFit benchmark workouts including:
- **Girl WODs**: Fran, Cindy, Diane, Helen, Grace, Isabel, Annie, Nancy, Karen, Linda, Nicole, Gwen, Hope, Candy, Margareta, Maggie, etc. (27 total Girl WODs)
- **Hero WODs**: Murph, DT, JT, Randy, Nate, Jason, Michael, Daniel, Tommy V, Josh, Luce, RJ, Whitten, Zeus, Ryan, Tiff, Holbrook, etc.
- **Benchmark WODs**: Filthy Fifty, Fight Gone Bad, King Kong, The Chief, The Ghost, Heavy Fran, Bear Complex

**CSV Structure:**
```
id,name,source,type,regime,score_type,description,url,notes,is_standard,created_by
```

**Field Descriptions:**
- `id`: Unique identifier (integer)
- `name`: WOD name (string)
- `source`: Origin of workout (e.g., "CrossFit")
- `type`: WOD category - `Girl`, `Hero`, `Benchmark`, `Games`, etc.
- `regime`: Workout format - `AMRAP`, `Fastest Time`, `EMOM`, etc.
- `score_type`: How workout is scored - `Time (HH:MM:SS)`, `Rounds+Reps`, `Max Weight`
- `description`: Full workout description with movements and rep schemes (string)
- `url`: Reference URL (optional, may be empty)
- `notes`: Additional information about the WOD (optional)
- `is_standard`: Always `TRUE` for standard WODs
- `created_by`: NULL for standard WODs (user ID for custom WODs)

## Usage

### Loading Seed Data on New Instance

#### Option 1: Database Import (Recommended for Production)

**For PostgreSQL:**
```bash
# Import movements
psql -d actalog -c "\COPY movements(id, name, description, type, is_standard, created_by) FROM 'seeds/movements.csv' WITH CSV HEADER;"

# Import WODs
psql -d actalog -c "\COPY wods(id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by) FROM 'seeds/wods.csv' WITH CSV HEADER;"

# Update sequences
psql -d actalog -c "SELECT setval('movements_id_seq', (SELECT MAX(id) FROM movements));"
psql -d actalog -c "SELECT setval('wods_id_seq', (SELECT MAX(id) FROM wods));"
```

**For MySQL:**
```sql
-- Import movements
LOAD DATA LOCAL INFILE 'seeds/movements.csv'
INTO TABLE movements
FIELDS TERMINATED BY ','
ENCLOSED BY '"'
LINES TERMINATED BY '\n'
IGNORE 1 ROWS
(id, name, description, type, is_standard, @created_by)
SET created_by = NULLIF(@created_by, '');

-- Import WODs
LOAD DATA LOCAL INFILE 'seeds/wods.csv'
INTO TABLE wods
FIELDS TERMINATED BY ','
ENCLOSED BY '"'
LINES TERMINATED BY '\n'
IGNORE 1 ROWS
(id, name, source, type, regime, score_type, description, @url, @notes, is_standard, @created_by)
SET url = NULLIF(@url, ''), notes = NULLIF(@notes, ''), created_by = NULLIF(@created_by, '');

-- Update auto increment
ALTER TABLE movements AUTO_INCREMENT = 96;
ALTER TABLE wods AUTO_INCREMENT = 69;
```

**For SQLite:**
```bash
# Import movements
sqlite3 actalog.db <<EOF
.mode csv
.import seeds/movements.csv movements
DELETE FROM movements WHERE id = 'id'; -- Remove header row if it got imported
EOF

# Import WODs
sqlite3 actalog.db <<EOF
.mode csv
.import seeds/wods.csv wods
DELETE FROM wods WHERE id = 'id'; -- Remove header row if it got imported
EOF
```

#### Option 2: API Import (For Application-Level Seeding)

Create a seeding script that reads CSV and uses API endpoints:

```bash
#!/bin/bash
# Example using curl to seed via API

# Get admin token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}' | jq -r '.token')

# Import movements
while IFS=, read -r id name description type is_standard created_by; do
  if [ "$id" != "id" ]; then
    curl -X POST http://localhost:8080/api/movements \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"name\":\"$name\",\"description\":\"$description\",\"type\":\"$type\"}"
  fi
done < seeds/movements.csv

# Import WODs (similar approach)
```

#### Option 3: Go Seeding Function

Add to `internal/repository/database.go`:

```go
func SeedStandardData(db *sqlx.DB) error {
    // Read and parse movements.csv
    movementsFile, _ := os.Open("seeds/movements.csv")
    movementsReader := csv.NewReader(movementsFile)
    movementsReader.Read() // Skip header

    for {
        record, err := movementsReader.Read()
        if err == io.EOF {
            break
        }
        // Parse and insert movement
        _, err = db.Exec(`
            INSERT INTO movements (name, description, type, is_standard)
            VALUES (?, ?, ?, ?)
            ON CONFLICT (name) DO NOTHING
        `, record[1], record[2], record[3], record[4])
    }

    // Similar for WODs...
    return nil
}
```

### Load Testing

Use these CSV files to generate load test data:

```bash
# Generate 1000 workout logs using random WODs
for i in {1..1000}; do
  WOD_ID=$((1 + RANDOM % 50))
  curl -X POST http://localhost:8080/api/user-workouts/log \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"workout_id\":$WOD_ID,\"workout_date\":\"2024-01-01\"}"
done
```

### User Import Templates

Users can create their own movements or WODs by following the CSV format:

**Custom Movement Template:**
```csv
id,name,description,type,is_standard,created_by
,My Custom Lift,Description of my movement,weightlifting,FALSE,<user_id>
```

**Custom WOD Template:**
```csv
id,name,source,type,regime,score_type,description,url,notes,is_standard,created_by
,My Custom WOD,Self-created,Self-created,AMRAP,Rounds+Reps,"20 min AMRAP: 10 Push-ups, 15 Squats",,Personal training workout,FALSE,<user_id>
```

## Data Notes

### Movement Types
- `weightlifting`: Barbell, dumbbell, kettlebell movements requiring external load
- `gymnastics`: Bodyweight movements requiring gymnastic skill (rings, bars, etc.)
- `bodyweight`: Basic bodyweight movements
- `cardio`: Monostructural cardio (row, run, bike, jump rope)

### WOD Types
- `Girl`: Original CrossFit female-named benchmarks (Fran, Cindy, etc.)
- `Hero`: Workouts honoring fallen military and first responders
- `Benchmark`: CrossFit.com benchmark workouts
- `Games`: CrossFit Games workouts
- `Self-created`: User custom WODs

### WOD Regimes
- `AMRAP`: As Many Rounds As Possible
- `Fastest Time`: For time (complete as fast as possible)
- `EMOM`: Every Minute On the Minute
- `Rounds+Reps`: Scored by rounds and reps completed
- `Max Weight`: Test maximum weight lifted

### Score Types
- `Time (HH:MM:SS)`: Workout completed for time (hours:minutes:seconds format)
- `Rounds+Reps`: Number of complete rounds plus additional reps
- `Max Weight`: Maximum weight achieved
- `Total Reps`: Total repetitions completed

## Timestamps

Note: The CSV files do NOT include `created_at` and `updated_at` timestamps. These should be automatically generated by the database or application when inserting records.

If your database requires explicit timestamps, add them during import:

```sql
-- PostgreSQL example
INSERT INTO movements (name, description, type, is_standard, created_at, updated_at)
VALUES ('Movement Name', 'Description', 'weightlifting', TRUE, NOW(), NOW());
```

## Maintenance

### Adding New Data

To add new movements or WODs:

1. Append new rows to the appropriate CSV file
2. Ensure IDs are sequential (next available ID)
3. Maintain consistent formatting
4. Set `is_standard` to `TRUE` for official CrossFit movements/WODs
5. Leave `created_by` empty for standard data

### Version History

- **v1.2** (2025-11-16): Expanded seed data
  - Added 20 new movements (76-95): Romanian Deadlift, Handstand Walk, Wall Walk, Hip Thrust, Box Step-up, Broad Jump, L-Sit, Hollow Hold, Arch Hold, Ring Row, various push-up variations, Medicine Ball Clean, Sandbag Carry, Yoke Carry, Atlas Stone Lift, Tire Flip, Battle Ropes, Airdyne Bike
  - Added 18 new WODs (51-68): Missing Girl WODs (Linda, Nicole, Gwen, Hope, Candy, Margareta, Maggie), additional Hero WODs (Josh, Luce, RJ, Whitten, Zeus, Ryan, Tiff, Holbrook), and Benchmark variations (Heavy Fran, King Kong For Time, Bear Complex)
  - Total: 95 standard movements and 68 famous WODs
  - Added Strongman category movements (Atlas Stone, Tire Flip, Yoke Carry)
  - Expanded gymnastics movements with static holds and variations
- **v1.1** (2025-01-16): Updated seed data
  - Fixed score_type format from `Time (MM:SS)` to `Time (HH:MM:SS)`
  - Removed duplicate "Power Snatch" entry
  - Corrected table name from `strength_movements` to `movements`
  - 74 standard movements (Olympic lifts, CrossFit movements)
  - 50 famous WODs (Girls, Heroes, Benchmarks)
- **v1.0** (2024-11-13): Initial seed data
  - 75 standard movements (Olympic lifts, CrossFit movements)
  - 50 famous WODs (Girls, Heroes, Benchmarks)

## References

- [CrossFit Benchmarks](https://www.crossfit.com/workouts)
- [CrossFit Hero WODs](https://www.crossfit.com/heroes/)
- [CrossFit Girls WODs](https://wodwell.com/wods/girls/)
- [Olympic Weightlifting](https://www.teamusa.org/usa-weightlifting)

## License

This seed data is compiled from publicly available CrossFit benchmarks and standard exercise nomenclature. CrossFit® is a registered trademark of CrossFit, LLC.

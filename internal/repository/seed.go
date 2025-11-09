package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// seedStandardMovements seeds the database with standard CrossFit movements
func seedStandardMovements(db *sql.DB) error {
	// Check if movements already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM movements WHERE is_standard = 1").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		fmt.Println("Standard movements already seeded, skipping...")
		return nil
	}

	fmt.Println("Seeding standard CrossFit movements...")

	now := time.Now()

	// Standard movements organized by type
	movements := []struct {
		Name        string
		Description string
		Type        string
	}{
		// Weightlifting
		{"Back Squat", "Barbell squat with bar on back", "weightlifting"},
		{"Front Squat", "Barbell squat with bar on front shoulders", "weightlifting"},
		{"Overhead Squat", "Squat with barbell held overhead", "weightlifting"},
		{"Deadlift", "Lift barbell from ground to standing position", "weightlifting"},
		{"Sumo Deadlift High Pull", "Wide stance deadlift pulling to chin", "weightlifting"},
		{"Clean", "Lift barbell from ground to shoulders", "weightlifting"},
		{"Power Clean", "Clean without full squat", "weightlifting"},
		{"Hang Clean", "Clean starting from hang position", "weightlifting"},
		{"Squat Clean", "Clean with full squat catch", "weightlifting"},
		{"Snatch", "Lift barbell from ground to overhead in one motion", "weightlifting"},
		{"Power Snatch", "Snatch without full squat", "weightlifting"},
		{"Hang Snatch", "Snatch starting from hang position", "weightlifting"},
		{"Clean and Jerk", "Clean followed by overhead jerk", "weightlifting"},
		{"Thruster", "Front squat to overhead press", "weightlifting"},
		{"Push Press", "Overhead press with leg drive", "weightlifting"},
		{"Push Jerk", "Overhead jerk with dip under", "weightlifting"},
		{"Split Jerk", "Overhead jerk with split stance", "weightlifting"},
		{"Bench Press", "Press barbell from chest while lying on bench", "weightlifting"},
		{"Overhead Press", "Strict press barbell from shoulders to overhead", "weightlifting"},
		{"Shoulder Press", "Overhead press variation", "weightlifting"},

		// Gymnastics
		{"Pull-up", "Pull body up to bar, chin over bar", "gymnastics"},
		{"Chest-to-Bar Pull-up", "Pull-up bringing chest to bar", "gymnastics"},
		{"Strict Pull-up", "Pull-up without kipping motion", "gymnastics"},
		{"Kipping Pull-up", "Pull-up using hip swing for momentum", "gymnastics"},
		{"Muscle-up", "Pull-up transitioning to dip above rings or bar", "gymnastics"},
		{"Bar Muscle-up", "Muscle-up performed on pull-up bar", "gymnastics"},
		{"Ring Muscle-up", "Muscle-up performed on gymnastic rings", "gymnastics"},
		{"Handstand Push-up", "Push-up performed in handstand position", "gymnastics"},
		{"Strict Handstand Push-up", "Handstand push-up without kipping", "gymnastics"},
		{"Kipping Handstand Push-up", "Handstand push-up with kipping motion", "gymnastics"},
		{"Dip", "Lower and press body between parallel bars or rings", "gymnastics"},
		{"Ring Dip", "Dip performed on gymnastic rings", "gymnastics"},
		{"Toes-to-Bar", "Hang from bar and bring toes to touch bar", "gymnastics"},
		{"Knees-to-Elbow", "Hang from bar and bring knees to elbows", "gymnastics"},
		{"L-Sit", "Hold body in L-shape while supported", "gymnastics"},
		{"Handstand Hold", "Hold inverted position on hands", "gymnastics"},
		{"Handstand Walk", "Walk on hands while inverted", "gymnastics"},
		{"Rope Climb", "Climb rope using arms and legs", "gymnastics"},

		// Bodyweight
		{"Push-up", "Press body up from prone position", "bodyweight"},
		{"Sit-up", "Raise torso from supine to sitting position", "bodyweight"},
		{"Air Squat", "Bodyweight squat", "bodyweight"},
		{"Burpee", "Squat thrust to plank, push-up, jump up", "bodyweight"},
		{"Box Jump", "Jump onto elevated platform", "bodyweight"},
		{"Step-up", "Step onto elevated platform", "bodyweight"},
		{"Lunge", "Step forward lowering back knee toward ground", "bodyweight"},
		{"Walking Lunge", "Lunge while moving forward", "bodyweight"},
		{"Jump Squat", "Squat with explosive jump", "bodyweight"},
		{"Pistol Squat", "Single-leg squat", "bodyweight"},
		{"Plank Hold", "Hold body in straight line on forearms", "bodyweight"},
		{"Hollow Hold", "Hold body in hollow position on back", "bodyweight"},
		{"Arch Hold", "Hold body in arched position on stomach", "bodyweight"},

		// Cardio
		{"Row", "Rowing machine for distance or calories", "cardio"},
		{"Run", "Running for distance or time", "cardio"},
		{"Bike", "Stationary bike for distance or calories", "cardio"},
		{"Ski Erg", "Ski ergometer for distance or calories", "cardio"},
		{"Assault Bike", "Air resistance bike for calories", "cardio"},
		{"Jump Rope", "Single or double under rope jumps", "cardio"},
		{"Double Under", "Jump rope passing twice under feet per jump", "cardio"},
		{"Single Under", "Jump rope passing once under feet per jump", "cardio"},
		{"Shuttle Run", "Sprint back and forth between two points", "cardio"},

		// Olympic Lifting Accessories
		{"Hang Power Clean", "Power clean from hang position", "weightlifting"},
		{"Hang Power Snatch", "Power snatch from hang position", "weightlifting"},
		{"Squat Snatch", "Snatch with full squat catch", "weightlifting"},
		{"Pause Squat", "Squat with pause at bottom", "weightlifting"},
		{"Box Squat", "Squat to box or bench", "weightlifting"},
		{"Good Morning", "Hip hinge with barbell on back", "weightlifting"},
		{"Romanian Deadlift", "Deadlift with straight legs", "weightlifting"},
		{"Kettlebell Swing", "Hip hinge swinging kettlebell", "weightlifting"},
		{"Turkish Get-up", "Rising from ground to standing with weight overhead", "weightlifting"},

		// Strongman
		{"Farmer Carry", "Walk carrying heavy weights in each hand", "weightlifting"},
		{"Sled Push", "Push weighted sled", "cardio"},
		{"Sled Pull", "Pull weighted sled", "cardio"},
		{"Yoke Carry", "Walk with weighted yoke on shoulders", "weightlifting"},

		// Core
		{"GHD Sit-up", "Sit-up on glute-ham developer", "gymnastics"},
		{"V-up", "Simultaneous leg and torso raise to V-shape", "bodyweight"},
		{"Russian Twist", "Seated torso rotation", "bodyweight"},
		{"AbMat Sit-up", "Sit-up with abdominal mat", "bodyweight"},

		// Accessory
		{"Wall Ball", "Squat and throw medicine ball to target", "weightlifting"},
		{"Medicine Ball Clean", "Clean with medicine ball", "weightlifting"},
		{"Dumbbell Snatch", "Snatch with single dumbbell", "weightlifting"},
		{"Dumbbell Thruster", "Thruster with dumbbells", "weightlifting"},
		{"Devil Press", "Burpee with dumbbell snatch", "weightlifting"},
	}

	// Insert movements
	query := `INSERT INTO movements (name, description, type, is_standard, created_at, updated_at) VALUES (?, ?, ?, 1, ?, ?)`

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range movements {
		_, err := stmt.Exec(m.Name, m.Description, m.Type, now, now)
		if err != nil {
			return fmt.Errorf("failed to seed movement %s: %w", m.Name, err)
		}
	}

	fmt.Printf("✓ Seeded %d standard movements\n", len(movements))
	return nil
}

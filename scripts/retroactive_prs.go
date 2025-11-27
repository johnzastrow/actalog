package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database connection
	db, err := sql.Open("sqlite3", "./actalog.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userWorkoutRepo := repository.NewUserWorkoutRepository(db)
	workoutRepo := repository.NewWorkoutRepository(db)
	workoutMovementRepo := repository.NewWorkoutMovementRepository(db)
	userWorkoutMovementRepo := repository.NewUserWorkoutMovementRepository(db)
	userWorkoutWODRepo := repository.NewUserWorkoutWODRepository(db)
	wodRepo := repository.NewWODRepository(db)

	// Initialize service
	userWorkoutService := service.NewUserWorkoutService(
		userWorkoutRepo,
		workoutRepo,
		workoutMovementRepo,
		userWorkoutMovementRepo,
		userWorkoutWODRepo,
		wodRepo,
	)

	// Run retroactive PR flagging for user ID 1
	userID := int64(1)
	fmt.Printf("Running retroactive PR flagging for user ID %d...\n", userID)

	movementPRCount, wodPRCount, err := userWorkoutService.RetroactivelyFlagPRs(userID)
	if err != nil {
		log.Fatalf("Failed to retroactively flag PRs: %v", err)
	}

	fmt.Printf("Success! Flagged %d movement PRs and %d WOD PRs\n", movementPRCount, wodPRCount)
}

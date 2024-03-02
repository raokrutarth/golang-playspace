package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Actor represents the actor that plays a character.
type Actor struct {
	ID   int64  `gorm:"id,primary_key"`
	Name string `gorm:"name"`
}

// Character is one character from the database.
type Character struct {
	ID      int64  `gorm:"id,primaryKey"`
	ActorID int64  `gorm:"actor_id"`
	Name    string `gorm:"name"`

	Actor Actor
}

// CharacterFilters are used to filter the results of a List query.
type CharacterFilters struct {
	// ActorID matches on the actor's ID.
	ActorID int64

	// ActorName does a case-insensitive partial match on the actor name.
	ActorName string

	// Name does a case-insensitive partial match on the character name.
	Name string

	// SceneNumber filters by the scene that the character appears in.
	SceneNumber int64
}

func main() {
	var err error
	err = godotenv.Load("dev.env")
	if err != nil {
		log.Fatal(err)
	}

	db, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&Character{})
	if err != nil {
		log.Fatal(err)
	}

	cs, err := ListCharacters(db, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(cs)
}

// Open returns a gorm.DB instance for the given sql.DB sqlite instance.
func Open() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=America/Los_Angeles",
		os.Getenv("COCKROACH_HOST"),
		os.Getenv("COCKROACH_USER"),
		os.Getenv("COCKROACH_PASSWORD"),
		os.Getenv("COCKROACH_DATABASE"),
		5432,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// ListCharacters searches for characters in the database.
//
// If filters is nil, all characters are returned. Otherwise, the results are
// filtered by the criteria in filters.
func ListCharacters(db *gorm.DB, filters *CharacterFilters) ([]*Character, error) {
	q := db

	if filters != nil {
		if filters.ActorID != 0 {
			q = q.Where("actor_id = ?", filters.ActorID)
		} else if filters.ActorName != "" {
			q = q.
				Joins("Actor").
				Where("LOWER(actor.name) LIKE ?", "%"+strings.ToLower(filters.ActorName)+"%")
		}

		if filters.Name != "" {
			q = q.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(filters.Name)+"%")
		}

		if filters.SceneNumber != 0 {
			q = q.
				Joins("INNER JOIN scene_characters ON scene_characters.character_id=characters.id").
				Where("scene_characters.scene_id = ?", filters.SceneNumber)
		}
	}

	var characters []*Character
	err := q.Find(&characters).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}

	return characters, nil
}

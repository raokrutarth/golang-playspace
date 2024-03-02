package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// ErrNotFound is returned when updating or deleting a character that does not
// exist in the database.
var ErrNotFound = errors.New("not found")

// Character is one character from the database.
type Character struct {
	ID      int64  `db:"id"`
	ActorID int64  `db:"actor_id"`
	Name    string `db:"name"`
}

// CharacterStore loads and updates characters_sqlx in the database.
type CharacterStore struct {
	dbx *sqlx.DB
}

func Open() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=America/Los_Angeles",
		os.Getenv("COCKROACH_HOST"),
		os.Getenv("COCKROACH_USER"),
		os.Getenv("COCKROACH_PASSWORD"),
		os.Getenv("COCKROACH_DATABASE"),
		5432,
	)
	return sql.Open("postgres", dsn)
}

// NewCharacterStore creates a new CharacterStore.
func NewCharacterStore(db *sql.DB) *CharacterStore {
	return &CharacterStore{dbx: sqlx.NewDb(db, "postgres")}
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	var err error
	ctx := context.Background()

	err = godotenv.Load("dev.env")
	if err != nil {
		log.Fatal(err)
	}

	db, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	charStore := NewCharacterStore(db)

	err = charStore.Store(ctx, &Character{
		ActorID: int64(42),
		Name:    "foo bar",
	})
	if err != nil {
		log.Fatal(err)
	}

	cs, err := charStore.List(ctx, nil)
	if err != nil {
		log.Fatal(cs)
	}
	log.Print(cs)
}

// Get loads a character from the database by ID.
//
// If no character is found, Get returns a nil Character and no error.
func (cs *CharacterStore) Get(ctx context.Context, id int64) (*Character, error) {
	var c Character
	err := cs.dbx.GetContext(ctx, &c, `SELECT id, actor_id, name FROM characters_sqlx WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &c, err
}

// Store saves a character to the database. If the character has an ID, it will
// be updated. Otherwise, it will be inserted and the ID will be set.
//
// If the character has an ID and it does not exist in the database, Store
// returns ErrNotFound.
func (cs *CharacterStore) Store(ctx context.Context, c *Character) error {
	if c.ID == 0 {
		return cs.insert(ctx, c)
	}

	return cs.update(ctx, c)
}

func (cs *CharacterStore) insert(ctx context.Context, c *Character) error {
	rows, err := cs.dbx.NamedQueryContext(
		ctx,
		`INSERT INTO characters_sqlx (actor_id, name) VALUES (:actor_id, :name) RETURNING id`,
		c,
	)
	if err != nil {
		return fmt.Errorf("insert character: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&c.ID)
		if err != nil {
			return fmt.Errorf("insert character: %w", err)
		}
	}

	return nil
}

func (cs *CharacterStore) update(ctx context.Context, c *Character) error {
	res, err := cs.dbx.NamedExecContext(
		ctx,
		`UPDATE characters_sqlx SET actor_id = :actor_id, name = :name WHERE id = :id`,
		c,
	)
	if err != nil {
		return fmt.Errorf("update character: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete removes a character from the database.
//
// If the character does not exist in the database, Delete returns ErrNotFound.
func (cs *CharacterStore) Delete(ctx context.Context, id int64) error {
	res, err := cs.dbx.ExecContext(ctx, `DELETE FROM characters_sqlx WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete character: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
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

// List searches for characters_sqlx in the database.
//
// If filters is nil, all characters_sqlx are returned. Otherwise, the results are
// filtered by the criteria in filters.
func (cs *CharacterStore) List(ctx context.Context, filters *CharacterFilters) ([]*Character, error) {
	var args []interface{}
	query := "SELECT c.id, c.actor_id, c.name FROM characters_sqlx c"
	joins := []string{}
	where := []string{}

	if filters != nil {
		if filters.ActorID != 0 {
			where = append(where, "c.actor_id = ?")
			args = append(args, filters.ActorID)
		} else if filters.ActorName != "" {
			joins = append(joins, "JOIN actors a ON a.id = c.actor_id")
			where = append(where, "LOWER(a.name) LIKE ?")
			args = append(args, "%"+strings.ToLower(filters.ActorName)+"%")
		}

		if filters.Name != "" {
			where = append(where, "LOWER(c.name) LIKE ?")
			args = append(args, "%"+strings.ToLower(filters.Name)+"%")
		}

		if filters.SceneNumber != 0 {
			joins = append(joins, "JOIN scene_characters_sqlx sc ON sc.character_id = c.id")
			where = append(where, "sc.scene_id = ?")
			args = append(args, filters.SceneNumber)
		}
	}

	if len(joins) > 0 {
		query += " " + strings.Join(joins, " ")
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	var characters []*Character
	err := cs.dbx.SelectContext(ctx, &characters, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list characters: %w", err)
	}

	return characters, nil
}

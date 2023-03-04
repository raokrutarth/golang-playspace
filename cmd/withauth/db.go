package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	db *gorm.DB
}

func NewPostgresDB(logger *zerolog.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=America/Los_Angeles",
		"db-dev",
		"app",
		os.Getenv("POSTGRES_PASSWORD"),
		"gops_db",
		5432,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&User{},
		&RangeTransaction{},
		&ExpandedTransaction{},
	)
	if err != nil {
		return nil, err
	}
	dbname := db.Migrator().CurrentDatabase()
	tables, _ := db.Migrator().GetTables()
	logger.Info().Strs("tables", tables).Msgf("connected to database %s", dbname)

	return &PostgresDB{db}, nil
}

func (r *PostgresDB) GetUser(username string) (*User, error) {
	return &User{}, nil
}
func (r *PostgresDB) CreateSignInSession(username string) (string, error) {
	return "", nil
}
func (r *PostgresDB) IsSignInTokenValid(username string, token string) (bool, error) {
	return false, nil
}
func (r *PostgresDB) DeleteSignInToken(username string) error {
	return nil
}
func (r *PostgresDB) GetRangeTransactions(userID, simulationID uuid.UUID) ([]RangeTransaction, error) {
	return []RangeTransaction{}, nil
}
func (r *PostgresDB) AddRangeTransaction(
	userID, simulationID uuid.UUID,
	incomeOrExpense, category, notes string,
	recurrenceEveryDays int,
	recurrenceStart, recurrenceEnd time.Time,
	amount float64,
) error {
	return nil
}
func (r *PostgresDB) DeleteRangeTransaction(userID, simulationID, rangeTransactionID uuid.UUID) error {
	return nil
}

func (r *PostgresDB) UpdateSimulationRange(
	userID, simulationID uuid.UUID,
	recurrenceStart time.Time,
	recurrenceEnd time.Time,
) error {
	return nil
}

func (r *PostgresDB) GetExpandedTransactions(userID, simulationID uuid.UUID) ([]ExpandedTransaction, error) {
	return []ExpandedTransaction{}, nil
}
func (r *PostgresDB) DeleteExpandedTransaction(userID, simulationID, expandedTransactionID uuid.UUID) error {
	return nil
}

// // CreateList creates a new list with the given name, returning its ID.
// func (r *PostgresDB) CreateList(name string) (string, error) {
// 	id := m.makeListID(10)
// 	// Generate time here because SQLite's CURRENT_TIMESTAMP only returns seconds.
// 	timeCreated := time.Now().In(time.UTC).Format(time.RFC3339Nano)
// 	_ := r.db.Exec("INSERT INTO lists (id, name, time_created) VALUES (?, ?, ?)",
// 		id, name, timeCreated)
// 	return id, err
// }

// var listIDChars = "bcdfghjklmnpqrstvwxyz" // just consonants to avoid spelling words

// // makeListID creates a new randomized list ID.
// func (r *PostgresDB) makeListID(n int) string {
// 	id := make([]byte, n)
// 	for i := 0; i < n; i++ {
// 		index := m.rnd.Intn(len(listIDChars))
// 		id[i] = listIDChars[index]
// 	}
// 	return string(id)
// }

// // DeleteList (soft) deletes the given list (its items actually remain
// // untouched). It's not an error if the list doesn't exist.
// func (r *PostgresDB) DeleteList(id string) error {
// 	_ := r.db.Exec("UPDATE lists SET time_deleted = CURRENT_TIMESTAMP WHERE id = ?", id)
// 	return err
// }

// // GetList fetches one list and returns it, or nil if not found.
// func (r *PostgresDB) GetList(id string) (*List, error) {
// 	row := r.db.QueryRow(`
// 		SELECT id, name
// 		FROM lists
// 		WHERE id = ? AND time_deleted IS NULL
// 		`, id)
// 	var list List
// 	err := row.Scan(&list.ID, &list.Name)
// 	if err == sql.ErrNoRows {
// 		return nil, nil
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	list.Items, err = m.getListItems(id)
// 	return &list, err
// }

// func (r *PostgresDB) getListItems(listID string) ([]*Item, error) {
// 	rows, err := r.db.Query(`
// 		SELECT id, description, done
// 		FROM items
// 		WHERE list_id = ? AND time_deleted IS NULL
// 		ORDER BY id
// 		`, listID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var items []*Item
// 	for rows.Next() {
// 		var item Item
// 		err = rows.Scan(&item.ID, &item.Description, &item.Done)
// 		if err != nil {
// 			return nil, err
// 		}
// 		items = append(items, &item)
// 	}
// 	return items, rows.Err()
// }

// // AddItem adds an item with the given description to a list, returning the
// // item ID.
// func (r *PostgresDB) AddItem(listID, description string) (string, error) {
// 	result, err := r.db.Exec("INSERT INTO items (list_id, description) VALUES (?, ?)",
// 		listID, description)
// 	if err != nil {
// 		return "", err
// 	}
// 	id, err := result.LastInsertId()
// 	if err != nil {
// 		return "", err
// 	}
// 	return strconv.Itoa(int(id)), nil
// }

// // UpdateDone updates the "done" flag of the given item in a list.
// func (r *PostgresDB) UpdateDone(listID, itemID string, done bool) error {
// 	_ := r.db.Exec("UPDATE items SET done = ? WHERE list_id = ? AND id = ?",
// 		done, listID, itemID)
// 	return err
// }

// // DeleteItem (soft) deletes the given item in a list.
// func (r *PostgresDB) DeleteItem(listID, itemID string) error {
// 	_ := r.db.Exec(`
// 			UPDATE items
// 			SET time_deleted = CURRENT_TIMESTAMP
// 			WHERE list_id = ? AND id = ?
// 		`, listID, itemID)
// 	return err
// }

// // CreateSignIn creates a new sign-in and returns its secure ID.
// func (r *PostgresDB) CreateSignInSession(username string) (string, error) {
// 	id := generateSignInToken()
// 	_ := r.db.Exec("INSERT INTO sign_ins (id) VALUES (?)", id)
// 	return id, err
// }

// // IsSignInValid reports whether the given sign-in ID is valid.
// func (r *PostgresDB) IsSignInValid(id string) (bool, error) {
// 	row := r.db.QueryRow(`
// 		SELECT 1
// 		FROM sign_ins
// 		WHERE id = ? AND time_created > DATETIME('NOW', '-90 DAYS')
// 		`, id)
// 	var dummy int
// 	err := row.Scan(&dummy)
// 	if err == sql.ErrNoRows {
// 		return false, nil
// 	}
// 	if err != nil {
// 		return false, err
// 	}
// 	return true, nil
// }

// // DeleteSignIn deletes the given sign-in. It's not an error if the sign-in
// // doesn't exist.
// func (r *PostgresDB) DeleteSignIn(id string) error {
// 	_ := r.db.Exec("DELETE FROM sign_ins WHERE id = ?", id)
// 	return err
// }

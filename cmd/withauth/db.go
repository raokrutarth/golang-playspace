package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresDB struct {
	db     *gorm.DB
	logger *zerolog.Logger
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

	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	return &PostgresDB{db, logger}, nil
}

func (r *PostgresDB) AddUser(ID uuid.UUID, username, passwordHash, passwordSalt string) error {
	newUser := &User{
		ID:           ID,
		Username:     username,
		PasswordHash: passwordHash,
		PasswordSalt: passwordSalt,
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(newUser).Error
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
	ID, userID, simulationID uuid.UUID,
	title, incomeOrExpense, category, notes string,
	recurrenceEveryDays int,
	recurrenceStart, recurrenceEnd time.Time,
	amount float64,
) error {
	transaction := &RangeTransaction{
		ID:                  ID,
		SimulationID:        simulationID,
		UserID:              userID,
		Title:               title,
		IncomeOrExpense:     incomeOrExpense,
		Category:            category,
		Notes:               notes,
		RecurrenceEveryDays: recurrenceEveryDays,
		RecurrenceStart:     recurrenceStart,
		RecurrenceEnd:       recurrenceEnd,
		Amount:              amount,
	}
	// TODO remove this caluse
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(transaction)
	return result.Error
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

func (r *PostgresDB) GetExpandedTransactions(
	// userID, simulationID uuid.UUID,
	rTxns []RangeTransaction,
) ([]ExpandedTransaction, error) {
	// Fetch the first range transaction

	r.logger.Info().Msg("adding range txns")
	for _, rtx := range rTxns {
		err := r.AddRangeTransaction(
			rtx.ID,
			rtx.UserID,
			rtx.SimulationID,
			rtx.Title,
			rtx.IncomeOrExpense,
			rtx.Category,
			"",
			rtx.RecurrenceEveryDays,
			rtx.RecurrenceStart,
			rtx.RecurrenceEnd,
			rtx.Amount,
		)
		if err != nil {
			r.logger.Err(err).Msg("failed to add seed range txn")
		}
	}

	r.logger.Info().Msg("querying range txns")
	// Query the ExpandedTransaction table using generate_series
	var expandedTransactions []ExpandedTransaction
	query := `
		SELECT *
		FROM (
			SELECT 
				uuid_generate_v4() as id, 
				id as range_transaction_id, 
				user_id, 
				title,
				generate_series(
					date_trunc('day', recurrence_start),
					date_trunc('day', recurrence_end),
					'1 day'::interval * recurrence_every_days
				)::date AS transaction_date,
				income_or_expense,
				category, 
				amount,
				NOW() as created_at, 
				NOW() as updated_at
			FROM 
				range_transactions
		) expanded_txns
		ORDER BY 
		expanded_txns.transaction_date ASC
	`
	if err := r.db.Raw(query).Scan(&expandedTransactions).Error; err != nil {
		log.Fatalf("Failed to fetch expanded transactions: %v", err)
	}

	return expandedTransactions, nil
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

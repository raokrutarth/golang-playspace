package main

import (
	"errors"
	"fmt"
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
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(newUser).Error
}

func (r *PostgresDB) GetUser(username string) (*User, error) {
	var user User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *PostgresDB) CreateSignInSession(username string) (string, error) {
	token := generateSecureToken(32)
	result := r.db.Model(&User{}).
		Where(
			"username = ?",
			username,
		).Update("login_session_token", token)
	if result.Error != nil {
		return "", result.Error
	}
	return token, nil
}

func (r *PostgresDB) IsSignInTokenValid(username string, token string) (bool, error) {
	var user User
	result := r.db.Where(
		"username = ? AND login_session_token = ?",
		username,
		token,
	).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (r *PostgresDB) DeleteSignInToken(username string) error {
	result := r.db.Model(&User{}).
		Where("username = ?", username).
		Update("login_session_token", "")
	return result.Error
}

func (r *PostgresDB) addExpandedTransactionsForRangeTransaction(rtx *RangeTransaction) error {
	// FIXME memory error & use transaction+rollback
	// add the respective expanded transactions
	var expandedTransactions []ExpandedTransaction
	query := `
		SELECT *
		FROM (
			SELECT 
				uuid_generate_v4() as id, 
				id as range_transaction_id, 
				user_id, 
				simulation_id,
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
			WHERE
				id = ?
		) expanded_txns
		ORDER BY 
			expanded_txns.transaction_date ASC
	`
	if err := r.db.Raw(query, rtx.ID).Scan(&expandedTransactions).Error; err != nil {
		return err
	}

	result := r.db.Clauses().
		Create(expandedTransactions)

	if result.Error != nil {
		return result.Error
	}
	r.logger.Info().Msgf("added %d expanded transactions for range txn %s",
		len(expandedTransactions), rtx.ID)
	return nil
}

func (r *PostgresDB) AddRangeTransaction(rtx *RangeTransaction) error {
	result := r.db.First(rtx, "id = ?", rtx.ID)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		r.logger.Warn().Msgf("ignoring insert of range txn with id %s because it exists", rtx.ID)
		return nil
	}
	result = r.db.Create(rtx)
	if result.Error != nil {
		return result.Error
	}
	return r.addExpandedTransactionsForRangeTransaction(rtx)
}

func (db *PostgresDB) UpdateRangeTransaction(rangeTransactionID uuid.UUID, newValue *RangeTransaction) error {
	tx := db.db.Begin()

	if err := tx.Error; err != nil {
		return err
	}

	var rangeTx RangeTransaction

	if err := tx.Where("id = ?", rangeTransactionID).First(&rangeTx).Error; err != nil {
		tx.Rollback()
		return err
	}

	rangeTx.Title = newValue.Title
	rangeTx.IncomeOrExpense = newValue.IncomeOrExpense
	rangeTx.Category = newValue.Category
	rangeTx.Notes = newValue.Notes
	rangeTx.RecurrenceEveryDays = newValue.RecurrenceEveryDays
	rangeTx.RecurrenceStart = newValue.RecurrenceStart
	rangeTx.RecurrenceEnd = newValue.RecurrenceEnd
	rangeTx.Amount = newValue.Amount

	if err := tx.Save(&rangeTx).Error; err != nil {
		tx.Rollback()
		return err
	}

	// remove old expanded transactions
	if err := tx.Where(
		"range_transaction_id = ?",
		rangeTransactionID,
	).Delete(&ExpandedTransaction{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// TODO use a transaction everywhere
	if tx.Commit().Error != nil {
		err := db.addExpandedTransactionsForRangeTransaction(&rangeTx)
		if err != nil {
			return err
		}
		return nil
	}
	return tx.Commit().Error
}

func (r *PostgresDB) DeleteRangeTransaction(userID, simulationID, rangeTransactionID uuid.UUID) error {
	tx := r.db.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Where(
		"id = ? AND user_id = ? AND simulation_id = ?",
		rangeTransactionID, userID, simulationID,
	).Delete(&RangeTransaction{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where(
		"range_transaction_id = ?",
		rangeTransactionID,
	).Delete(&ExpandedTransaction{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *PostgresDB) ListRangeTransactions(userID, simulationID uuid.UUID) ([]RangeTransaction, error) {
	var rangeTransactions []RangeTransaction
	result := r.db.Where("user_id = ? AND simulation_id = ?", userID, simulationID).
		Find(&rangeTransactions).
		Order("updated_at DESC")
	if result.Error != nil {
		return nil, result.Error
	}
	return rangeTransactions, nil
}

func (r *PostgresDB) AddExpandedTransaction(etx *ExpandedTransaction) error {
	return r.db.Save(etx).Error
}

func (r *PostgresDB) UpdateExpandedTransaction(expandedTransactionID uuid.UUID, newValue *ExpandedTransaction) error {
	result := r.db.Model(&ExpandedTransaction{}).
		Where("id = ?", expandedTransactionID).
		Updates(map[string]interface{}{
			"title":             newValue.Title,
			"transaction_date":  newValue.TransactionDate,
			"income_or_expense": newValue.IncomeOrExpense,
			"category":          newValue.Category,
			"amount":            newValue.Amount,
			"updated_at":        time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no record updated")
	}
	return nil
}

func (db *PostgresDB) DeleteExpandedTransaction(userID, simulationID, expandedTransactionID uuid.UUID) error {
	result := db.db.Where(`id = ?`, expandedTransactionID).
		Delete(&ExpandedTransaction{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no record deleted")
	}
	return nil
}

func (db *PostgresDB) ListExpandedTransactions(userID, simulationID uuid.UUID) ([]ExpandedTransaction, error) {
	var transactions []ExpandedTransaction
	result := db.db.Where("user_id = ? AND simulation_id = ?", userID, simulationID).
		Find(&transactions)

	if result.Error != nil {
		return nil, result.Error
	}
	return transactions, nil
}

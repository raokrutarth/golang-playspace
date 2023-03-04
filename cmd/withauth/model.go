package main

import (
	"time"

	"github.com/gofrs/uuid"
)

//
// DB models
//

type User struct {
	ID                uuid.UUID `gorm:"primarykey"`
	Username          string    `gorm:"index;unique"`
	PasswordHash      string
	PasswordSalt      string
	LoginSessionToken string `gorm:"index"`
	RangeTransactions []RangeTransaction
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type RangeTransaction struct {
	ID           string    `gorm:"primarykey"`
	SimulationID uuid.UUID `gorm:"index"`
	UserID       uuid.UUID `gorm:"index"`

	IncomeOrExpense      string
	Category             string
	Notes                string
	RecurrenceEveryDays  int
	RecurrenceStart      time.Time
	RecurrenceEnd        time.Time
	Amount               float64
	ExpandedTransactions []ExpandedTransaction `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ExpandedTransaction struct {
	ID              uuid.UUID `gorm:"primarykey"`
	SimulationID    uuid.UUID `gorm:"index"`
	UserID          uuid.UUID
	Title           string
	TransactionDate time.Time
	IncomeOrExpense string
	Category        string

	CreatedAt time.Time
	UpdatedAt time.Time
}

//
// Frontend models
//

type SegmentedTransaction struct {
	ExpandedTransactionID uuid.UUID
	Title                 string
	TransactionDate       time.Time
	IncomeOrExpense       string
	Category              string
}

// all the data required to render the web page
type WebpageState struct {
	CSRFToken         string
	LoginSessionToken string

	SimulationID uuid.UUID

	Username string
	UserId   uuid.UUID

	RangeTransactions     []RangeTransaction
	SegmentedTransactions []SegmentedTransaction
}

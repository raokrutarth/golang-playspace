package main

import (
	"time"

	"github.com/gofrs/uuid"
)

//
// DB models
//

type User struct {
	ID                   uuid.UUID `gorm:"primarykey"`
	Username             string    `gorm:"index;unique"`
	PasswordHash         string
	PasswordSalt         string
	LoginSessionToken    string                `gorm:"index"`
	RangeTransactions    []RangeTransaction    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ExpandedTransactions []ExpandedTransaction `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type RangeTransaction struct {
	ID        uuid.UUID `gorm:"primarykey"`
	PlannerID uuid.UUID `gorm:"index"`
	UserID    uuid.UUID // FK

	Title               string
	IncomeOrExpense     string
	Category            string
	Notes               string
	RecurrenceEveryDays int
	RecurrenceStart     time.Time
	RecurrenceEnd       time.Time
	Amount              float64
	Source              string // bank/planner/bank-modified/card/brokerage
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ExpandedTransaction struct {
	ID                 uuid.UUID `gorm:"primarykey"`
	RangeTransactionID uuid.UUID
	UserID             uuid.UUID `gorm:"index"` // FK
	PlannerID          uuid.UUID `gorm:"index"`
	Title              string
	TransactionDate    time.Time
	IncomeOrExpense    string
	Category           string
	Amount             float64

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
	Amount                float64
	NetCash               float64
}

// all the data required to render the web page
type HomePageState struct {
	CSRFToken   string
	IsLoggedIn  bool
	SignInError bool

	// redirect url
	ReturnURL string

	PlannerID  uuid.UUID
	PlannerEnd time.Time

	RangeStart time.Time
	RangeEnd   time.Time

	Username string
	UserID   uuid.UUID

	RangeTransactions     []RangeTransaction
	SegmentedTransactions []*SegmentedTransaction
}

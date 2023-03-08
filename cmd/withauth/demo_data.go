package main

import (
	"time"

	"github.com/gofrs/uuid"
)

var bankRangeTxns = []RangeTransaction{
	{
		ID:                  uuid.Nil,
		SimulationID:        uuid.Nil,
		UserID:              uuid.Nil,
		Title:               "MyHome Rental Ltd.",
		IncomeOrExpense:     "expense",
		Category:            "Rent",
		Notes:               "",
		RecurrenceEveryDays: 7,
		RecurrenceStart:     time.Now(),
		RecurrenceEnd:       time.Now().AddDate(0, 0, 30),
		Amount:              750.0,
		Source:              "bank",
	},
	{
		ID:                  uuid.Nil,
		SimulationID:        uuid.Nil,
		UserID:              uuid.Nil,
		Title:               "GE paycheck ACH",
		IncomeOrExpense:     "income",
		Category:            "Salary",
		Notes:               "Received monthly salary",
		RecurrenceEveryDays: 30,
		RecurrenceStart:     time.Now().AddDate(0, -1, 0),
		RecurrenceEnd:       time.Now().AddDate(0, 3, 0),
		Amount:              5000.0,
		Source:              "bank",
	},
	{
		ID:                  uuid.Nil,
		SimulationID:        uuid.Nil,
		UserID:              uuid.Nil,
		Title:               "Side Business Investment",
		IncomeOrExpense:     "expense",
		Category:            "Business expenses",
		Notes:               "expenses to start side business",
		RecurrenceEveryDays: 15,
		RecurrenceStart:     time.Now().AddDate(0, -1, 0),
		RecurrenceEnd:       time.Now().AddDate(0, 7, 0),
		Amount:              725.0,
		Source:              "card",
	},
}

package main

import (
	"time"

	"github.com/gofrs/uuid"
)

var (
	u1, _ = uuid.FromString("e83d1ea6-eaae-4446-afed-ba8763674fe8")
	u2, _ = uuid.FromString("c61410f3-f044-4f02-a30c-84227b9b5f0c")
	u3, _ = uuid.FromString("1ab04292-57fd-4eb3-b431-39d4086035e9")
	u4, _ = uuid.FromString("dbce5f1a-4e26-486e-ae22-d1232a16f29a")
)

var bankRangeTxns = []RangeTransaction{
	{
		ID:                  u1,
		PlannerID:           uuid.Nil,
		UserID:              uuid.Nil,
		Title:               "MyHome Rental Ltd.",
		IncomeOrExpense:     "expense",
		Category:            "Rent",
		Notes:               "",
		RecurrenceEveryDays: 30,
		RecurrenceStart:     time.Now(),
		RecurrenceEnd:       time.Now().AddDate(0, 3, 0),
		Amount:              750.0,
		Source:              "bank",
	},
	{
		ID:                  u2,
		PlannerID:           uuid.Nil,
		UserID:              uuid.Nil,
		Title:               "GE paycheck ACH",
		IncomeOrExpense:     "income",
		Category:            "Salary",
		Notes:               "Received monthly salary",
		RecurrenceEveryDays: 30,
		RecurrenceStart:     time.Now().AddDate(0, -1, 0),
		RecurrenceEnd:       time.Now().AddDate(0, 3, 0),
		Amount:              1000.0,
		Source:              "bank",
	},
	{
		ID:                  u3,
		PlannerID:           uuid.Nil,
		UserID:              uuid.Nil,
		Title:               "Side Business Investment",
		IncomeOrExpense:     "expense",
		Category:            "Business expenses",
		Notes:               "expenses to start side business",
		RecurrenceEveryDays: 15,
		RecurrenceStart:     time.Now(),
		RecurrenceEnd:       time.Now().AddDate(0, 12, 0),
		Amount:              250.0,
		Source:              "card",
	},
}

var bankOneTimeTxns = []ExpandedTransaction{
	{
		ID:              u4,
		PlannerID:       uuid.Nil,
		UserID:          uuid.Nil,
		Title:           "Starting Balance",
		IncomeOrExpense: "income",
		TransactionDate: time.Now().AddDate(0, -6, 0),
		Amount:          750.0,
	},
}

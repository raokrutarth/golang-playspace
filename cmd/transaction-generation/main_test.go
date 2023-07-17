package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var (
	startDate           = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate             = time.Date(2053, 12, 31, 0, 0, 0, 0, time.UTC)
	transactionPatterns = generateMockTransactionPatterns(2000)
)

// generateMockTransactionPatterns generates up to n random TransactionPatterns with corner cases
func generateMockTransactionPatterns(n int) []TransactionPattern {
	mts := make([]TransactionPattern, 0, n)

	for i := 0; i < n; i++ {
		// Generate a random TransactionPattern
		txnDate := generateRandomDate()
		transactionPattern := TransactionPattern{
			Amount:      rand.Intn(1000) - 500, // Random amount between -500 and 499
			CategoryID:  fmt.Sprintf("category_%d", i),
			Date:        txnDate.Format("2006-01-02"),
			Description: fmt.Sprintf("description_%d", i),
			ID:          fmt.Sprintf("TransactionPattern_%d", i),
			IsRecurring: rand.Float32() < 0.5, // 50% chance of being recurring
			Title:       fmt.Sprintf("title_%d", i),
			Type:        "income",
		}

		// Set recurrence pattern for recurring TransactionPatterns
		if transactionPattern.IsRecurring {
			transactionPattern.RecurrencePattern.StartDate = transactionPattern.Date
			transactionPattern.RecurrencePattern.EndDate = txnDate.Add(
				time.Duration(rand.Intn(270000)+24) * time.Hour,
			).Format("2006-01-02")
			transactionPattern.RecurrencePattern.FrequencyHours = rand.Intn(730) + 1
			transactionPattern.RecurrencePattern.FrequencyType = "custom"
		}
		// Append the TransactionPattern to the list
		mts = append(mts, transactionPattern)
	}
	return mts
}

// generateRandomDate generates a random date between 2010-01-01 and 2023-12-31
func generateRandomDate() time.Time {
	days := int(endDate.Sub(startDate).Hours() / 24)
	randomDays := rand.Intn(days)
	return startDate.Add(time.Duration(randomDays) * 24 * time.Hour)
}

func TestGenerateTimeSeriesPoints(t *testing.T) {
	r := generateTimeSeriesPoints(transactionPatterns)
	if len(r) < 20000*30 {
		t.Fatal("only generated: ", len(r), " with ", len(transactionPatterns))
	}
	t.Log("number of individual txns: ", len(r))
}

func BenchmarkGenerateTimeSeriesPoints(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateTimeSeriesPoints(transactionPatterns)
	}
}

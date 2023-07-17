package main

import (
	"log"
	"time"
)

type TransactionPattern struct {
	Amount            int               `json:"amount"`
	CategoryID        string            `json:"category_id"`
	Date              string            `json:"date"`
	Description       string            `json:"description"`
	ID                string            `json:"id"`
	IsRecurring       bool              `json:"is_recurring"`
	RecurrencePattern RecurrencePattern `json:"recurrence_pattern"`
	Title             string            `json:"title"`
	Type              string            `json:"type"`
}

type RecurrencePattern struct {
	EndDate        string `json:"end_date"`
	FrequencyHours int    `json:"frequency_hours"`
	FrequencyType  string `json:"frequency_type"`
	IsTimeBound    bool   `json:"is_time_bound"`
	StartDate      string `json:"start_date"`
}

type TimeSeriesDataPoint struct {
	Date   string `json:"date"`
	Amount int    `json:"amount"`
	// Add additional fields as needed
}

func main() {
	log.Println("hello, world")
}

func generateTimeSeriesPoints(transactionPatterns []TransactionPattern) []TimeSeriesDataPoint {
	totalPoints := 0
	for _, tp := range transactionPatterns {
		if tp.IsRecurring {
			recurrencePattern := tp.RecurrencePattern
			startDate, _ := time.Parse("2006-01-02", recurrencePattern.StartDate)
			endDate, _ := time.Parse("2006-01-02", recurrencePattern.EndDate)
			duration := endDate.Sub(startDate)
			frequencyDuration := time.Duration(recurrencePattern.FrequencyHours) * time.Hour
			points := int(duration / frequencyDuration)
			totalPoints += points
		} else {
			totalPoints++
		}
	}

	timeSeriesPoints := make([]TimeSeriesDataPoint, 0, totalPoints)

	// Generate time series points
	for _, tp := range transactionPatterns {
		if tp.IsRecurring {
			recurrencePattern := tp.RecurrencePattern
			startDate, _ := time.Parse("2006-01-02", recurrencePattern.StartDate)
			endDate, _ := time.Parse("2006-01-02", recurrencePattern.EndDate)
			duration := endDate.Sub(startDate)
			frequencyDuration := time.Duration(recurrencePattern.FrequencyHours) * time.Hour
			points := int(duration / frequencyDuration)

			currentDate := startDate
			for i := 0; i < points; i++ {
				timeSeriesPoints = append(timeSeriesPoints, TimeSeriesDataPoint{
					Date:   currentDate.Format("2006-01-02"),
					Amount: tp.Amount,
				})
				currentDate = currentDate.Add(frequencyDuration)
			}
		} else {
			timeSeriesPoints = append(timeSeriesPoints, TimeSeriesDataPoint{
				Date:   tp.Date,
				Amount: tp.Amount,
			})
		}
	}
	return timeSeriesPoints
}

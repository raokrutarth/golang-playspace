package main

func (rt *RangeTransaction) Expand() []ExpandedTransaction {

	// in-memory computation of
	numDays := int(rt.RecurrenceEnd.Sub(rt.RecurrenceStart).Hours() / 24)
	if numDays == 0 {
		numDays = 1
	}
	amountPerDay := rt.Amount / float64(numDays)

	var expanded []ExpandedTransaction
	for i := 0; i < numDays; i++ {
		day := rt.RecurrenceStart.AddDate(0, 0, i)
		var incomeOrExpense string
		if rt.IncomeOrExpense == "income" {
			incomeOrExpense = "income"
		} else {
			incomeOrExpense = "expense"
		}
		expanded = append(expanded, ExpandedTransaction{
			RangeTransactionID: rt.ID,
			Title:              rt.Title,
			TransactionDate:    day,
			IncomeOrExpense:    incomeOrExpense,
			Category:           rt.Category,
			Amount:             amountPerDay,
		})
	}
	return expanded
}

# C GPT

## db implementation

```go
type Repository interface {
 GetUser(username string) (*User, error)
 CreateSignInSession(username string) (string, error)
 IsSignInTokenValid(username string, token string) (bool, error)
 DeleteSignInToken(username string) error

 GetRangeTransactions(userID, simulationID uuid.UUID) ([]RangeTransaction, error)
 AddRangeTransaction(
  userID, simulationID uuid.UUID,
  incomeOrExpense, category, notes string,
  recurrenceEveryDays int,
  recurrenceStart, recurrenceEnd time.Time,
  amount float64,
 ) error
 DeleteRangeTransaction(userID, simulationID, rangeTransactionID uuid.UUID) error

 UpdateSimulationRange(
  userID, simulationID uuid.UUID,
  recurrenceStart time.Time,
  recurrenceEnd time.Time,
 ) error

 GetExpandedTransactions(userID, simulationID uuid.UUID) ([]ExpandedTransaction, error)
 DeleteExpandedTransaction(userID, simulationID, expandedTransactionID uuid.UUID) error
}


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
 ID           string    `gorm:"primarykey"`
 SimulationID uuid.UUID `gorm:"index"`
 UserID       uuid.UUID // FK

 Title                string
 IncomeOrExpense      string
 Category             string
 Notes                string
 RecurrenceEveryDays  int
 RecurrenceStart      time.Time
 RecurrenceEnd        time.Time
 Amount               float64
 ExpandedTransactions []ExpandedTransaction `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
 CreatedAt            time.Time
 UpdatedAt            time.Time
}

type ExpandedTransaction struct {
 ID                 uuid.UUID `gorm:"primarykey"`
 RangeTransactionID uuid.UUID // FK
 UserID             uuid.UUID // FK
 Title              string
 TransactionDate    time.Time
 IncomeOrExpense    string
 Category           string

 CreatedAt time.Time
 UpdatedAt time.Time
}
```

for the interface and models above, as struct PostgresDB that implements the Repository interface is:

## demo data gen

```sql
create temporary table simulation_txns
  (title varchar, transaction_date TIMESTAMPTZ, );

insert into simulation_txns values <range transaction examples here>;

select
    productNumber,
    generate_series(firstYearAvailable,lastYearAvailable) as productYear
  from simulation_txns;
```

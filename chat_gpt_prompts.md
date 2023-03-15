# ChatGPT

## db implementation

```go
type Repository interface {
 AddUser(ID uuid.UUID, username, passwordHash, passwordSalt string) error
 GetUser(username string) (*User, error)
 CreateSignInSession(username string) (string, error)
 IsSignInTokenValid(username string, token string) (bool, error)
 DeleteSignInToken(username string) error

 AddRangeTransaction(rtx *RangeTransaction) error
 UpdateRangeTransaction(rangeTransactionID uuid.UUID, newValue *RangeTransaction) error
 DeleteRangeTransaction(userID, simulationID, rangeTransactionID uuid.UUID) error
 ListRangeTransactions(userID, simulationID uuid.UUID) ([]RangeTransaction, error)

 AddExpandedTransaction(etx *ExpandedTransaction) error
 UpdateExpandedTransaction(expandedTransactionID uuid.UUID, newValue *ExpandedTransaction) error
 DeleteExpandedTransaction(userID, simulationID, expandedTransactionID uuid.UUID) error
 ListExpandedTransactions(userID, simulationID uuid.UUID) ([]ExpandedTransaction, error)
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

for the interface and models above, a struct PostgresDB that implements the Repository interface would implement DeleteRangeTransaction as

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

## in-memory expanded tx gen

```go

type RangeTransaction struct {
  ID           uuid.UUID `gorm:"primarykey"`
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
  Source               string                // bank/simulation/bank-modified/card/brokerage
  ExpandedTransactions []ExpandedTransaction `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
  CreatedAt            time.Time
  UpdatedAt            time.Time
}

type ExpandedTransaction struct {
  ID                 uuid.UUID `gorm:"primarykey"`
  RangeTransactionID uuid.UUID // FK
  Title              string
  TransactionDate    time.Time
  IncomeOrExpense    string
  Category           string
  Amount             float64

  CreatedAt time.Time
  UpdatedAt time.Time
}
```

for the structs above, the golang code that converts a range transaction to an array of ExpandedTransaction structs represnting individual expenses or incomes is:

## simulations struct with json cols

define a struct `Simulations` representing a gorm model that has the columns UserID as a foreign key from the user table, RangeTransactions containing a json array of range transactions, ItemizedTransactions containing an json array of ExpandedTransactions computed by the Expand() function above, and the usual columns like id, createdAt, UpdatedAt and SoftDeleteAt

```go
type Simulations struct {
  ID                   uuid.UUID `gorm:"primarykey"`
  UserID               uuid.UUID // FK
  RangeTransactions    []RangeTransaction `gorm:"type:json"`
  ItemizedTransactions []ExpandedTransaction `gorm:"type:json"`
  CreatedAt            time.Time
  UpdatedAt            time.Time
  DeletedAt            gorm.DeletedAt `gorm:"index"`
}
```

```go
type PostgresDB struct {
 db     *gorm.DB
 logger *zerolog.Logger
}
```

given the struct above, implement the CRUDL operations for the Simulations struct including CRUDL for the individual transactions columns

## nav bar fix

```html
<nav>
    <div class="nav-wrapper">
        <a href="#!" class="brand-logo">Proto Inc.</a>
        <ul class="right hide-on-med-and-down">
            <li><a href="#">Home</a></li>
            <li>
                <a class="dropdown-trigger" href="#!" data-target="dropdown2">
                    Accounts<i class="material-icons right">arrow_drop_down</i>
                </a>
            </li>
            <!-- Dropdown Trigger -->
            <li>
                <a class="dropdown-trigger" href="#!" data-target="dropdown1">
                    Simulations<i class="material-icons right">arrow_drop_down</i>
                </a>
            </li>
        </ul>
    </div>
</nav>
```

fix the above materializecss nav bar so the drop downs don't open on top on the nav bar

## chat card creation

```html
<div class="card">
    <div class="card-content">
        <span class="card-title">Form 3</span>
        <form action="/add-free-flow" method="POST" enctype="application/x-www-form-urlencoded">
            <div class="input-field">
                <textarea id="textarea1" class="materialize-textarea" placeholder="rent costs $300 every month"></textarea>
                <label for="textarea1">Add a hypothetical income/expense</label>
            </div>
            <button class="btn waves-effect waves-light" type="submit" name="action">Submit</button>
        </form>
    </div>
</div>
```

convert the card above to a chat like form with a chat history containing a message from another sender and ability to scroll to previous messages

## clear icon

```html
<a href="#!" class="secondary-content"><i class="material-icons">clear</i></a>
```

convert the icon to red delete icon and an edit button with spacing in between

## add simulation to nav bar

```html
<div class="row">
    <div class="col s5">
        <ul class="collapsible">
            <li>
                <div class="collapsible-header">
                    <i class="material-icons">show_chart</i>
                    Simulation: <strong>{{ .SimulationID }}</strong>
                    <span class="new badge red" data-badge-caption="risk">4</span>
                    <span class="new badge blue" data-badge-caption="opportunity">4</span>
                </div>
                <div class="collapsible-body">
                    <p>Home purchase planning. <a href="#"><i class="material-icons">edit</i></a></p>
                    <p class="blue">$10k investable savings in July.</p>
                    <p class="red">Low balance of $50 in April</p>

                </div>
            </li>
        </ul>
    </div>
</div>
```

this collapsible component with rounded edges, a shadow, the body items as pills

## icon size

```html
<td class="left">
                <a href="#!" style="margin-left: 0px;">
                    <i class="tiny material-icons blue-text darken-4">edit</i>
                </a>
                <a>
                    <i class="tiny material-icons red-text darken-4">delete</i>
                </a>
            </td>
```

to make sure the icons are always side-by-side the change would be

## fileinput in chat card

```html
<div class="card chat-card">
                <div class="card-content chat-card-content">
                    <span class="card-title">Free Flow Chat</span>
                    <div class="chat-history">
                        <div class="chat-message sent">
                            <p class="chat-message-text">spend $30 every month on gym membership</p>
                        </div>
                        <div class="chat-message received">
                            <p class="chat-message-text">When does this $30 expense begin and end?</p>
                        </div>
                        <div class="chat-message sent">
                            <p class="chat-message-text">december and 6 months after</p>
                        </div>
                        <div class="chat-message received">
                            <p class="chat-message-text">Got it!</p>
                        </div>
                        <!-- Add more messages here -->
                    </div>
                    <form action="/add-free-flow" method="POST" enctype="application/x-www-form-urlencoded" class="chat-form">
                        <div class="input-field">
                            <textarea name="message" id="chat-message" class="materialize-textarea"></textarea>
                            <label for="chat-message">Type your message</label>
                        </div>
                        <button type="submit" class="btn waves-effect waves-light">Send</button>
                    </form>
                    
                </div>
            </div>
```

use MaterializeCSS to create form with a text area input with a send icon and file upload icon within the test input pill. similar to how the whatsapp chat input looks. use the clip icon for the file attachment and show the filename below the text input and give it a x to remove the file before sending

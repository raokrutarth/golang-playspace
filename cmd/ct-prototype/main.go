package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"golang.org/x/term"
)

const demoSimulationID = "22850c5a-eeee-4995-9eb0-1f6897acdc7e"

func main() {
	err := godotenv.Load("dev.env")
	exitOnError(err)

	addUser := flag.Bool("add-user", false, "-")
	flag.Parse()

	log := NewConsole(false)

	repository, err := NewPostgresDB(log)
	exitOnError(err)

	// use the application as a cli tool to add an admin user to the db
	if *addUser {
		var password string
		for len(password) < 6 {
			fmt.Printf("Enter password (at least 6-16 chars): ")
			b, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			exitOnError(err)
			password = string(b)
		}
		newuserID, _ := uuid.NewV4()
		salt := generateSecureToken(8)
		hash, err := GeneratePasswordHash(password, salt)
		exitOnError(err)
		err = repository.AddUser(
			newuserID,
			os.Getenv("ADMIN_USERNAME"),
			hash,
			salt,
		)
		exitOnError(err)
		log.Info().Msgf("added user %s with id %s", os.Getenv("ADMIN_USERNAME"), newuserID)
		return
	}

	// adminUsername, ok := os.LookupEnv("ADMIN_USERNAME")
	// if !ok {
	// 	exitOnError(errors.New("admin username not set"))
	// }
	// adminPasswordHash, ok := os.LookupEnv("ADMIN_PASSHASH")
	// if !ok {
	// 	exitOnError(errors.New("admin password hash not set"))
	// }

	server, err := NewServer(
		repository,
		log,
		"",
		"",
	)
	exitOnError(err)

	port := 5000
	if portEnv, ok := os.LookupEnv("PORT"); ok {
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			exitOnError(err)
		}
	}
	log.Info().Msgf("Serving traffic on port %d", port)
	err = http.ListenAndServe(":"+strconv.Itoa(port), server)
	exitOnError(err)
}

func NewConsole(isDebug bool) *zerolog.Logger {
	logLevel := zerolog.InfoLevel
	if isDebug {
		logLevel = zerolog.TraceLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	return &logger
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Server is the HTTP server for the to-do list app.
type Server struct {
	repository Repository
	logger     *zerolog.Logger

	// admin creds
	AdminUsername     string
	AdminPasswordHash string

	mux      *http.ServeMux
	homeTmpl *template.Template
	listTmpl *template.Template
}

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

func NewServer(
	repository Repository,
	logger *zerolog.Logger,
	username string,
	passwordHash string,
) (*Server, error) {
	s := &Server{
		repository: repository,
		logger:     logger,
		mux:        http.NewServeMux(),
	}
	s.addRoutes()
	return s, nil
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Cache-Control", "no-cache")
	s.mux.ServeHTTP(w, r)
	s.logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Dur("response-time", time.Since(startTime)).Send()
}

func (s *Server) addRoutes() {
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" { // because "/" pattern matches /*
			s.home(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	// s.mux.HandleFunc("/healthz", healthz())
	s.mux.HandleFunc("/sign-in", csrf(s.signIn))
	s.mux.HandleFunc("/sign-out", s.signedIn(csrf(s.signOut)))

	s.mux.HandleFunc("/demo", s.seedDemoData)

	s.mux.HandleFunc("/add-range-transaction", s.signedIn(csrf(s.addRangeEntry)))
	s.mux.HandleFunc("/update-range-transaction", s.signedIn(csrf(s.notImplemented)))
	s.mux.HandleFunc("/delete-range-transaction", s.signedIn(csrf(s.deleteRangeEntry)))

	s.mux.HandleFunc("/add-one-time-transaction", s.signedIn(csrf(s.addOneTimeEntry)))
	s.mux.HandleFunc("/update-one-time-transaction", s.signedIn(csrf(s.notImplemented)))
	s.mux.HandleFunc("/delete-one-time-transaction", s.signedIn(csrf(s.deleteOneTimeEntry)))

	s.mux.HandleFunc("/add-free-flow", s.signedIn(csrf(s.notImplemented)))
}

func (s *Server) signIn(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	returnURL := r.FormValue("return-url")
	if returnURL == "" {
		returnURL = "/"
	}
	s.logger.Info().Msgf(
		"sign in attempt with user %s and return url %s",
		username, returnURL,
	)
	user, err := s.repository.GetUser(username)
	if err != nil {
		s.internalError(w, "unable to get user", err)
		return
	}
	s.logger.Info().Msgf("got user with id %s", user.ID)

	// TODO rate-limit
	if os.Getenv("BYPASS_LOGIN") != "true" || CheckPasswordHash(
		user.PasswordSalt,
		password,
		user.PasswordHash,
	) != nil {
		s.logger.Info().Str("pw_bypass", os.Getenv("BYPASS_LOGIN")).Msgf("password verification failed")
		location := "/?error=sign-in&return-url=" + url.QueryEscape(returnURL)
		http.Redirect(w, r, location, http.StatusFound)
		return
	}
	token, err := s.repository.CreateSignInSession(username)
	if err != nil {
		s.internalError(w, "creating sign in", err)
		return
	}
	initBrowserSession(w, r, username, token)

	s.logger.Info().Msgf("login success for user %s", username)
	http.Redirect(w, r, returnURL, http.StatusFound)
}

func (s *Server) signOut(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "session_token",
		MaxAge:   -1,
		Path:     "/",
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	userLogin, err := extractUserLogin(r)
	if err != nil {
		s.internalError(w, "deleting sign in", err)
		return
	}

	err = s.repository.DeleteSignInToken(
		userLogin.Username,
	)
	if err != nil {
		s.internalError(w, "deleting sign in", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) signedIn(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.isSignedIn(r) {
			location := "/?return-url=" + url.QueryEscape(r.URL.Path)
			http.Redirect(w, r, location, http.StatusFound)
			return
		}
		h(w, r)
	}
}

func (s *Server) isSignedIn(r *http.Request) bool {
	userLogin, err := extractUserLogin(r)
	if err != nil {
		return false
	}

	// TODO rate-limit
	valid, err := s.repository.IsSignInTokenValid(
		userLogin.Username,
		userLogin.SessionToken,
	)
	return err == nil && valid
}

func (s *Server) seedDemoData(w http.ResponseWriter, r *http.Request) {
	var err error
	simulationID, _ := uuid.FromString(demoSimulationID)

	isSignedIn := s.isSignedIn(r)
	if !isSignedIn {
		newuserID, _ := uuid.NewV4()
		salt := generateSecureToken(8)
		hash, err := GeneratePasswordHash(os.Getenv("ADMIN_PASSWORD"), salt)
		exitOnError(err)
		err = s.repository.AddUser(
			newuserID,
			os.Getenv("ADMIN_USERNAME"),
			hash,
			salt,
		)
		if err != nil {
			s.internalError(w, "unable to add seed user", err)
			return
		}
	}

	user, _ := s.repository.GetUser(os.Getenv("ADMIN_USERNAME"))

	for _, rtx := range bankRangeTxns {
		rtx.UserID = user.ID
		rtx.SimulationID = simulationID
		err = s.repository.AddRangeTransaction(&rtx)
		if err != nil {
			s.internalError(w, "unable to add range seed data", err)
			return
		}
	}

	for _, etx := range bankOneTimeTxns {
		etx.UserID = user.ID
		etx.SimulationID = simulationID
		err = s.repository.AddExpandedTransaction(&etx)
		if err != nil {
			s.internalError(w, "unable to add one time seed data", err)
			return
		}
	}

	s.logger.Info().Msgf("added %d range transactins and %d one time entries", len(bankRangeTxns), len(bankOneTimeTxns))
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	isSignedIn := s.isSignedIn(r)
	simulationID, err := uuid.FromString(demoSimulationID)
	if err != nil {
		s.internalError(w, "unable to render template", err)
		return
	}

	var allRangeTxns []RangeTransaction
	var segTxns []*SegmentedTransaction
	if isSignedIn {
		userLogin, err := extractUserLogin(r)
		if err != nil {
			s.internalError(w, "unable to get user login", err)
			return
		}
		user, _ := s.repository.GetUser(userLogin.Username)

		rangeTxns, err := s.repository.ListRangeTransactions(user.ID, simulationID)
		if err != nil {
			s.internalError(w, "unable to fetch range txns", err)
			return
		}

		expandedTransactions, _ := s.repository.ListExpandedTransactions(
			user.ID, simulationID,
		)
		for _, etx := range expandedTransactions {
			segTxns = append(segTxns, &SegmentedTransaction{
				ExpandedTransactionID: etx.ID,
				Title:                 etx.Title,
				TransactionDate:       etx.TransactionDate,
				IncomeOrExpense:       etx.IncomeOrExpense,
				Amount:                etx.Amount,
			})
		}

		sort.SliceStable(segTxns, func(i, j int) bool {
			return segTxns[i].TransactionDate.Before(segTxns[j].TransactionDate)
		})
		netCash := float64(0)
		for _, stx := range segTxns {
			if stx.IncomeOrExpense == "income" {
				netCash += stx.Amount
			} else {
				netCash -= stx.Amount
			}
			stx.NetCash = netCash
		}

		allRangeTxns = rangeTxns
	}

	data := HomePageState{
		CSRFToken:             getCSRFToken(w, r),
		IsLoggedIn:            isSignedIn,
		ReturnURL:             r.URL.Query().Get("return-url"),
		SignInError:           r.URL.Query().Get("error") == "sign-in",
		SimulationID:          simulationID,
		SimulationEnd:         time.Now().AddDate(1, 0, 0),
		RangeStart:            time.Now(),
		RangeEnd:              time.Now().AddDate(0, 1, 0),
		Username:              os.Getenv("ADMIN_USERNAME"),
		UserID:                uuid.Nil,
		RangeTransactions:     allRangeTxns,
		SegmentedTransactions: segTxns,
	}

	// render the login screen
	s.logger.Info().Msg("rendering base template")
	if err := StaticResources.ExecuteTemplate(w, "index.html", data); err != nil {
		s.internalError(w, "unable to render template", err)
	}
}

func (s *Server) addRangeEntry(w http.ResponseWriter, r *http.Request) {
	userLogin, err := extractUserLogin(r)
	if err != nil {
		s.internalError(w, "unable to get user login", err)
		return
	}
	user, _ := s.repository.GetUser(userLogin.Username)

	simulationID, err := uuid.FromString(demoSimulationID)
	if err != nil {
		s.internalError(w, "unable to render template", err)
		return
	}
	newUUID, _ := uuid.NewV4()

	if err := r.ParseForm(); err != nil {
		s.internalError(w, "unable to parse form", err)
		return
	}

	_ = func(timeInput string) time.Time {
		t, err := time.Parse("Jan 02, 2006", timeInput)
		if err != nil {
			s.internalError(w, "unable parse time input", err)
		}
		return t
	}

	validate := validator.New()
	type Form struct {
		Title           string  `form:"title" validate:"required,min=0,max=255"`
		IncomeOrExpense string  `form:"income_or_expense" validate:"required,lowercase,min=0,max=255"`
		Category        string  `form:"category" validate:"min=0,max=255"`
		Notes           string  `form:"notes" validate:"min=0,max=255"`
		Amount          float64 `form:"amount" validate:"required,gt=0"`

		RecurrenceEveryDays int       `form:"recurrence_every" validate:"required"`
		RecurrenceStart     time.Time `form:"recurrence_start"`
		RecurrenceEnd       time.Time `form:"recurrence_end"`
	}
	// validate:"datetime='Jan 02, 2006'"

	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	var form Form
	err = decoder.Decode(&form, r.PostForm)
	if err != nil {
		s.internalError(w, "unable to parse POST form", err)
		l := s.logger.Error()
		for k, v := range r.PostForm {
			l = l.Strs(k, v)
		}
		l.Send()
		return
	}
	err = validate.Struct(form)
	if err != nil {
		s.internalError(w, "unable to validate POST form", err)
		return
	}
	if time.Now().AddDate(0, 0, -1).After(form.RecurrenceStart) {
		fmt.Fprintf(w, "<h2>cannot have a future item starting in the past. focus on the future</h2>")
		return
	}
	if form.RecurrenceEnd.Before(form.RecurrenceStart) {
		s.internalError(w, "recurrence cannot end before it starts", err)
		return
	}

	transaction := &RangeTransaction{
		ID:                  newUUID,
		SimulationID:        simulationID,
		UserID:              user.ID,
		Title:               form.Title,
		IncomeOrExpense:     form.IncomeOrExpense,
		Category:            form.Category,
		Notes:               form.Notes,
		RecurrenceEveryDays: form.RecurrenceEveryDays,
		RecurrenceStart:     form.RecurrenceStart,
		RecurrenceEnd:       form.RecurrenceEnd,
		Amount:              form.Amount,
		Source:              "simulation",
	}

	if err = s.repository.AddRangeTransaction(transaction); err != nil {
		s.internalError(w, "unable to save range tnx", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) deleteRangeEntry(w http.ResponseWriter, r *http.Request) {
	userLogin, err := extractUserLogin(r)
	if err != nil {
		s.internalError(w, "unable to get user login", err)
		return
	}
	user, _ := s.repository.GetUser(userLogin.Username)
	simulationID, err := uuid.FromString(demoSimulationID)
	if err != nil {
		s.internalError(w, "unable to render template", err)
		return
	}

	if err := r.ParseForm(); err != nil {
		s.internalError(w, "unable to parse form", err)
		return
	}

	validate := validator.New()
	type Form struct {
		RangeTransactionID string `form:"range_transaction_id" validate:"required,uuid4"`
	}

	decoder := form.NewDecoder()
	var form Form
	err = decoder.Decode(&form, r.PostForm)
	if err != nil {
		s.internalError(w, "unable to parse POST form", err)
		l := s.logger.Error()
		for k, v := range r.PostForm {
			l = l.Strs(k, v)
		}
		l.Send()
		return
	}
	err = validate.Struct(form)
	if err != nil {
		s.internalError(w, "unable to validate POST form", err)
		return
	}
	id, _ := uuid.FromString(form.RangeTransactionID)
	err = s.repository.DeleteRangeTransaction(user.ID, simulationID, id)
	if err != nil {
		s.internalError(w, "unable to delete range transaction", err)
		return
	}
	s.logger.Info().Msgf("deleted range transaction with id %s", form.RangeTransactionID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) addOneTimeEntry(w http.ResponseWriter, r *http.Request) {
	userLogin, err := extractUserLogin(r)
	if err != nil {
		s.internalError(w, "unable to get user login", err)
		return
	}
	user, _ := s.repository.GetUser(userLogin.Username)
	simulationID, err := uuid.FromString(demoSimulationID)
	if err != nil {
		s.internalError(w, "unable to render template", err)
		return
	}

	if err := r.ParseForm(); err != nil {
		s.internalError(w, "unable to parse form", err)
		return
	}

	validate := validator.New()
	type Form struct {
		Title           string    `form:"title" validate:"required,min=1,max=255"`
		IncomeOrExpense string    `form:"income_or_expense" validate:"required,oneof=income expense"`
		Category        string    `form:"category" validate:"min=0,max=255"`
		Amount          float64   `form:"amount" validate:"required,gt=0"`
		TransactionDate time.Time `form:"transaction_date"`
	}

	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	var form Form
	err = decoder.Decode(&form, r.PostForm)
	if err != nil {
		s.internalError(w, "unable to parse POST form", err)
		l := s.logger.Error()
		for k, v := range r.PostForm {
			l = l.Strs(k, v)
		}
		l.Send()
		return
	}
	err = validate.Struct(form)
	if err != nil {
		s.internalError(w, "unable to validate POST form", err)
		return
	}
	newUUID, _ := uuid.NewV4()
	err = s.repository.AddExpandedTransaction(&ExpandedTransaction{
		ID:                 newUUID,
		UserID:             user.ID,
		RangeTransactionID: uuid.Nil,
		SimulationID:       simulationID,
		Title:              form.Title,
		IncomeOrExpense:    form.IncomeOrExpense,
		Category:           form.Category,
		TransactionDate:    form.TransactionDate,
		Amount:             form.Amount,
	})
	if err != nil {
		s.internalError(w, "unable to add item", err)
		return
	}
	s.logger.Info().Msgf("added one-time transaction with id %s", newUUID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) deleteOneTimeEntry(w http.ResponseWriter, r *http.Request) {
	userLogin, err := extractUserLogin(r)
	if err != nil {
		s.internalError(w, "unable to get user login", err)
		return
	}
	user, _ := s.repository.GetUser(userLogin.Username)
	simulationID, err := uuid.FromString(demoSimulationID)
	if err != nil {
		s.internalError(w, "unable to render template", err)
		return
	}

	if err := r.ParseForm(); err != nil {
		s.internalError(w, "unable to parse form", err)
		return
	}

	validate := validator.New()
	type Form struct {
		ExpandedTransactionID string `form:"expanded_transaction_id" validate:"required,uuid4"`
	}

	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	var form Form
	err = decoder.Decode(&form, r.PostForm)
	if err != nil {
		s.internalError(w, "unable to parse POST form", err)
		l := s.logger.Error()
		for k, v := range r.PostForm {
			l = l.Strs(k, v)
		}
		l.Send()
		return
	}
	err = validate.Struct(form)
	if err != nil {
		s.internalError(w, "unable to validate POST form", err)
		return
	}
	id, _ := uuid.FromString(form.ExpandedTransactionID)
	err = s.repository.DeleteExpandedTransaction(user.ID, simulationID, id)
	if err != nil {
		s.internalError(w, "unable to add item", err)
		return
	}
	s.logger.Info().Msgf("added one-time transaction with id %s", id)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) notImplemented(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>not implemented</h1> [TODO] add metric.")
}

func (s *Server) internalError(w http.ResponseWriter, msg string, err error) {
	s.logger.Err(err).Msgf("Returning internal error with message %s", msg)
	http.Error(w, "error: "+msg, http.StatusInternalServerError)
}

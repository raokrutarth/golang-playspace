// Tiny to-do list web app

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/joho/godotenv"
	"github.com/raokrutarth/golang-playspace/templates"
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

	GetRangeTransactions(userID, simulationID uuid.UUID) ([]RangeTransaction, error)
	AddRangeTransaction(rtx *RangeTransaction) error
	DeleteRangeTransaction(userID, simulationID, rangeTransactionID uuid.UUID) error

	UpdateSimulationRange(
		userID, simulationID uuid.UUID,
		recurrenceStart time.Time,
		recurrenceEnd time.Time,
	) error

	GetExpandedTransactions(rTxns []RangeTransaction) ([]ExpandedTransaction, error)
	DeleteExpandedTransaction(userID, simulationID, expandedTransactionID uuid.UUID) error
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

	s.mux.HandleFunc("/add-range-transaction", s.signedIn(csrf(s.addRangeEntry)))
	s.mux.HandleFunc("/delete-range-transaction", s.signedIn(csrf(s.deleteList)))
	s.mux.HandleFunc("/update-simulation-range", s.signedIn(csrf(s.addItem)))
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

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Cache-Control", "no-cache")
	s.mux.ServeHTTP(w, r)
	s.logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(startTime))
}

func (s *Server) renderSimulation() WebpageState {

	uuidSample1, err := uuid.FromString(
		"dd94434f-ed0b-4837-ac63-205da7dae1c0",
	)
	if err != nil {
		s.logger.Fatal().Msgf("uer uuid search failed with error %s", err)
	}
	uuidSample2, _ := uuid.NewV4()
	uuidSample3, _ := uuid.NewV4()
	uuidSample4, _ := uuid.NewV4()
	uuidSample5, _ := uuid.NewV4()

	s.repository.AddUser(
		uuidSample1,
		"demo-manual-del-me@t.com",
		"foobar",
		"abcd",
	)

	rangeTxns := []RangeTransaction{
		{
			ID:                  uuidSample3,
			SimulationID:        uuidSample2,
			UserID:              uuidSample1,
			Title:               "Walmart San Jose/CA",
			IncomeOrExpense:     "expense",
			Category:            "Personal Expenses",
			Notes:               "Bought groceries",
			RecurrenceEveryDays: 7,
			RecurrenceStart:     time.Now(),
			RecurrenceEnd:       time.Now().AddDate(0, 0, 30),
			Amount:              50.0,
		},
		{
			ID:                  uuidSample4,
			SimulationID:        uuidSample2,
			UserID:              uuidSample1,
			Title:               "GE paycheck ACH",
			IncomeOrExpense:     "income",
			Category:            "Salary",
			Notes:               "Received monthly salary",
			RecurrenceEveryDays: 30,
			RecurrenceStart:     time.Now().AddDate(0, -1, 0),
			RecurrenceEnd:       time.Now().AddDate(0, 3, 0),
			Amount:              5000.0,
		},
		{
			ID:                  uuidSample5,
			SimulationID:        uuidSample2,
			UserID:              uuidSample1,
			Title:               "Side Business Investment",
			IncomeOrExpense:     "expense",
			Category:            "Business expenses",
			Notes:               "expenses to start side business",
			RecurrenceEveryDays: 15,
			RecurrenceStart:     time.Now().AddDate(0, -1, 0),
			RecurrenceEnd:       time.Now().AddDate(0, 7, 0),
			Amount:              725.0,
		},
	}

	expTxns, _ := s.repository.GetExpandedTransactions(rangeTxns)

	segTxns := []SegmentedTransaction{}
	netCash := float64(0)
	for _, tx := range expTxns {
		if tx.IncomeOrExpense == "income" {
			netCash += tx.Amount
		} else {
			netCash -= tx.Amount
		}
		segTxns = append(segTxns, SegmentedTransaction{
			Title:           tx.Title,
			TransactionDate: tx.TransactionDate,
			IncomeOrExpense: tx.IncomeOrExpense,
			Amount:          tx.Amount,
			NetCash:         netCash,
		})
	}
	s.logger.Info().Msgf("first exptx: %+v", expTxns[0])

	// s.logger.Info().Msgf("last exptx: %+v", expTxns[len(expTxns)-1])

	data := WebpageState{
		CSRFToken:             "",
		LoginSessionToken:     "",
		SimulationID:          uuidSample2,
		SimulationEnd:         time.Now().AddDate(1, 0, 0),
		RangeStart:            time.Now(),
		RangeEnd:              time.Now().AddDate(0, 1, 0),
		Username:              "demo-user",
		UserID:                uuidSample1,
		RangeTransactions:     rangeTxns,
		SegmentedTransactions: segTxns,
	}
	return data
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	isSignedIn := s.isSignedIn(r)
	simulationID, err := uuid.FromString(demoSimulationID)
	if err != nil {
		s.internalError(w, "unable to render template", err)
		return
	}

	var allRangeTxns []RangeTransaction
	var segTxns []SegmentedTransaction
	if isSignedIn {
		userLogin, err := extractUserLogin(r)
		if err != nil {
			s.internalError(w, "unable to get user login", err)
			return
		}
		user, _ := s.repository.GetUser(userLogin.Username)

		rangeTxns, err := s.repository.GetRangeTransactions(user.ID, simulationID)
		if err != nil {
			s.internalError(w, "unable to fetch range txns", err)
		}

		netCash := float64(0)
		for _, rtx := range rangeTxns {
			for _, etx := range rtx.ExpandedTransactions {
				if etx.IncomeOrExpense == "income" {
					netCash += etx.Amount
				} else {
					netCash -= etx.Amount
				}
				segTxns = append(segTxns, SegmentedTransaction{
					Title:           etx.Title,
					TransactionDate: etx.TransactionDate,
					IncomeOrExpense: etx.IncomeOrExpense,
					Amount:          etx.Amount,
					NetCash:         netCash,
				})
			}

		}
		allRangeTxns = append(rangeTxns, bankRangeTxns...)
	}

	data := WebpageState{
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
	if err := templates.Resources.ExecuteTemplate(w, "index.html", data); err != nil {
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
		Title           string  `form:"title" validate:"required,alphanum,min=0,max=255"`
		IncomeOrExpense string  `form:"income_or_expense" validate:"required,lowercase,min=0,max=255"`
		Category        string  `form:"category" validate:"lowercase,min=0,max=255"`
		Notes           string  `form:"notes" validate:"min=0,max=255"`
		Amount          float64 `form:"amount" validate:"required,gt=0"`

		RecurrenceEveryDays int       `form:"recurrence_every" validate:"required"`
		RecurrenceStart     time.Time `form:"recurrence_start"`
		RecurrenceEnd       time.Time `form:"recurrence_end"`
	}

	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.Parse("Jan 02, 2006", vals[0])
	}, time.Time{})

	var form Form
	err = decoder.Decode(&form, r.PostForm)
	if err != nil {
		s.internalError(w, "unable to parse POST form", err)
	}
	err = validate.Struct(form)
	if err != nil {
		s.internalError(w, "unable to validate POST form", err)
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

func (s *Server) showList(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/simulation/"):]
	s.logger.Info().Msgf("%d", id)
	// uuid, _ := uuid.NewV4()
	list, err := s.repository.GetExpandedTransactions([]RangeTransaction{})
	if err != nil {
		s.internalError(w, "fetching list", err)
		return
	}
	if list == nil {
		http.NotFound(w, r)
		return
	}

	data := struct {
		Prefill struct {
			Now       string
			NowPlus5y string
		}
	}{
		Prefill: struct {
			Now       string
			NowPlus5y string
		}{time.Now().Format("Jan 02, 2006"), time.Now().Add(time.Hour * 24 * 365).Format("Jan 02, 2006")},
	}
	fmt.Printf("rendering data %+v\n", data)
	if err := templates.Resources.ExecuteTemplate(w, "index.html", data); err != nil {
		s.internalError(w, "rendering template", err)
	}
}

func (s *Server) createList(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		// Empty list name, just reload home page
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	listID := "foo"
	// listID, err := s.repository.CreateList(name)
	// if err != nil {
	// 	s.internalError(w, "creating list", err)
	// 	return
	// }
	http.Redirect(w, r, "/simulation/"+listID, http.StatusFound)
}

func (s *Server) deleteList(w http.ResponseWriter, r *http.Request) {
	// id := r.FormValue("list-id")
	uuid, _ := uuid.NewV4()
	err := s.repository.DeleteRangeTransaction(uuid, uuid, uuid)
	if err != nil {
		s.internalError(w, "deleting list", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) addItem(w http.ResponseWriter, r *http.Request) {
	listID := r.FormValue("list-id")
	list, err := s.repository.GetUser(listID)
	if err != nil {
		s.internalError(w, "fetching list", err)
		return
	}
	if list == nil {
		http.NotFound(w, r)
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	if description == "" {
		// Empty item description, just reload list
		http.Redirect(w, r, "/simulation/"+list.Username, http.StatusFound)
		return
	}
	_, err = s.repository.CreateSignInSession(list.Username)
	if err != nil {
		s.internalError(w, "adding item", err)
		return
	}
	http.Redirect(w, r, "/simulation/"+list.Username, http.StatusFound)
}

func (s *Server) updateDone(w http.ResponseWriter, r *http.Request) {
	listID := r.FormValue("list-id")
	// itemID := r.FormValue("item-id")
	// done := r.FormValue("done") == "on"
	// err := s.repository.UpdateDone(listID, itemID, done)
	// if err != nil {
	// 	s.internalError(w, "updating done flag", err)
	// 	return
	// }
	http.Redirect(w, r, "/simulation/"+listID, http.StatusFound)
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) {
	listID := r.FormValue("list-id")
	// itemID := r.FormValue("item-id")
	// err := s.repository.DeleteItem(listID, itemID)
	// if err != nil {
	// 	s.internalError(w, "deleting item", err)
	// 	return
	// }
	http.Redirect(w, r, "/simulation/"+listID, http.StatusFound)
}

func (s *Server) internalError(w http.ResponseWriter, msg string, err error) {
	s.logger.Err(err).Msgf("Returning internal error with message %s", msg)
	http.Error(w, "error "+msg, http.StatusInternalServerError)
}

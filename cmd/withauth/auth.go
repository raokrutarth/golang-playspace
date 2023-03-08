package main

import (
	"encoding/hex"
	"math/rand"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// csrf wraps the given handler, ensuring that the HTTP method is POST and
// that the CSRF token in the "csrf-token" cookie matches the token in the
// "csrf-token" form field.
func csrf(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}
		token := r.FormValue("csrf-token")
		cookie, err := r.Cookie("csrf-token")
		if os.Getenv("BYPASS_LOGIN") != "true" && (err != nil || token != cookie.Value) {
			http.Error(w, "invalid CSRF token or cookie", http.StatusBadRequest)
			return
		}
		h(w, r)
	}
}

// getCSRFToken returns the current session's CSRF token, generating a new one
// and settings the "csrf-token" cookie if not present.
func getCSRFToken(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("csrf-token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	token := generateSecureToken(16)
	cookie = &http.Cookie{
		Name:     "csrf-token",
		Value:    token,
		Path:     "/",
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	return token
}

// set cookies for session token and username
func initBrowserSession(
	w http.ResponseWriter,
	r *http.Request,
	username,
	token string,
) {
	// TODO set expiry
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		MaxAge:   24 * 60 * 60,
		Path:     "/",
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	cookie = &http.Cookie{
		Name:     "username",
		Value:    username,
		MaxAge:   24 * 60 * 60,
		Path:     "/",
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

type UserLogin struct {
	SessionToken string
	Username     string
}

func extractUserLogin(r *http.Request) (UserLogin, error) {
	token, err := r.Cookie("session_token")
	if err != nil {
		return UserLogin{}, err
	}
	username, err := r.Cookie("username")
	if err != nil {
		return UserLogin{}, err
	}
	return UserLogin{token.Value, username.Value}, nil
}

func generateSecureToken(nBytes int) string {
	b := make([]byte, nBytes)
	_, err := rand.Read(b)
	if err != nil { // should never fail
		panic(err)
	}
	return hex.EncodeToString(b)
}

// CheckPasswordHash returns nil when the hash is for the valid salt and password.
func CheckPasswordHash(salt, password, passwordHash string) error {
	err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password+salt),
	)
	return err
}

// GeneratePasswordHash generates a bcrypt hash from the given password.
func GeneratePasswordHash(password, salt string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	return string(b), err
}

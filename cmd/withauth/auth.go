package main

import (
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"

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
		if err != nil || token != cookie.Value {
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
	token := generateCSRFToken()
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

func generateSignInToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil { // should never fail
		panic(err)
	}
	return hex.EncodeToString(b)
}

func getSignInCookie(r *http.Request) string {
	cookie, err := r.Cookie("sign-in")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func generateCSRFToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil { // should never fail
		panic(err)
	}
	return hex.EncodeToString(b)
}

// CheckPasswordHash returns a non-nil error if the given password hash is not
// a valid bcrypt hash.
func CheckPasswordHash(passwordHash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("x"))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil
	}
	return err
}

// GeneratePasswordHash generates a bcrypt hash from the given password.
func GeneratePasswordHash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

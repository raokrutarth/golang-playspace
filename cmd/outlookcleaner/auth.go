package main

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/raokrutarth/golang-playspace/pkg/logger"
	"golang.org/x/term"
)

func AuthValidate(ctx context.Context) {
	l := logger.GetLoggerFromContext(ctx)
	var err error
	connections, err := NewMailAccountConnections(ctx)
	if err != nil {
		l.Error("failed to get account connections", "error", err)
		return
	}
	for _, c := range connections {
		if err = c.client.Logout(); err != nil {
			l.Error("failed logout", "error", err)
		}
	}
}

// TODO
// https://codereview.stackexchange.com/questions/125846/encrypting-strings-in-golang

func encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// Encrypt method is to encrypt or hide any classified text
func encrypt(text, key, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, []byte(iv))
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return encode(cipherText), nil
}

func promptForCredentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ") // permit
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Password: ") //permit
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

func ReadAndInitCredentials(ctx context.Context) {
	cfg := getConfig(ctx).Encrypt
	l := logger.GetLoggerFromContext(ctx)
	encSecret := cfg.Secret
	l.Info("F\found configured encryption secret and initialization vector.",
		"secretLen", len(cfg.Secret), "ivLen", len(cfg.Iv))

	username, password, _ := promptForCredentials()
	l.Info("Encrypting username and password", "uname-len", len(username))

	encText, err := encrypt(username, encSecret, cfg.Iv)
	if err != nil {
		l.Error("Unable to encrypt password with error.", "error", err)
		return
	}
	fmt.Printf("Encrypted username: %s\n", encText)

	encText, err = encrypt(password, encSecret, cfg.Iv)
	if err != nil {
		l.Error("Unable to encrypt password with error.", "error", err)
		os.Exit(1)
	}
	fmt.Printf("Encrypted password: %s\n", encText)
	l.Info("Add the encrypted credentials above to the app config.")
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// Decrypt method is to extract back the encrypted text
func Decrypt(text, key, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	cipherText := Decode(text)
	cfb := cipher.NewCFBDecrypter(block, []byte(iv))
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

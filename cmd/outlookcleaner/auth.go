package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

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

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ") // permit
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Password: ") //permit
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

func ReadAndInitCredentials() {
	cfg := getConfig().Encrypt
	encSecret := cfg.Secret
	log.Info().
		Int("secret-len", len(cfg.Secret)).
		Int("iv-len", len(cfg.Iv)).
		Msg("Found configured encryption secret and initialization vector.")

	username, password, _ := credentials()
	log.Info().Int("uname-len", len(username)).Msg("Encrypting username and password.")

	encText, err := encrypt(username, encSecret, cfg.Iv)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to encrypt password with error.")
	}
	log.Info().Msg("Encrypted username: " + encText)

	encText, err = encrypt(password, encSecret, cfg.Iv)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to encrypt password with error.")
	}
	log.Info().Msg("Encrypted password: " + encText)

	log.Warn().Msg("Add the encrypted credentials above to the app config.")
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

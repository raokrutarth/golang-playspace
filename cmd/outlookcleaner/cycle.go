package main

// import (
// 	"fmt"
// 	"os"
// 	"strings"

// 	"github.com/emersion/go-imap"
// 	"github.com/emersion/go-imap/client"
// 	"github.com/raokrutarth/golang-playspace/cmd/outlook-cleaner/internal/config"
// 	"github.com/raokrutarth/golang-playspace/cmd/outlook-cleaner/internal/credstore"
// 	"github.com/rs/zerolog/log"
// 	"github.com/samber/lo"
// )

// func Cycle() error {
// 	log.Info().Msg("Running ingest cycle.")

// 	var err error
// 	var username, password string

// 	decryptionKey := config.GlobalConfig.Encrypt.Secret
// 	decryptionIv := config.GlobalConfig.Encrypt.Iv

// 	for i, account := range *config.GlobalConfig.Mail.Accounts {
// 		log.Info().Int("account_idx", i).Str("username", account.EncUser).Msg("Ingesting emails")

// 		username, err = credstore.Decrypt(account.EncUser, decryptionKey, decryptionIv)
// 		if err != nil {
// 			log.Err(err).Msg("Unable to decrypt username with error. Skipping account.")
// 			continue
// 		}

// 		password, err = credstore.Decrypt(account.EncPassword, decryptionKey, decryptionIv)
// 		if err != nil {
// 			log.Err(err).Msg("Unable to decrypt password with error. Skipping account.")
// 			continue
// 		}

// 		imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", account.Hostname, account.Port), nil)

// 		if err != nil {
// 			log.Err(err).Str("hostname", account.Hostname).Msg("Unable to connect to mail server.")
// 			continue
// 		}

// 		log.Info().Str("hostname", account.Hostname).Msg("Successfully connected to mail server")

// 		if err := imapClient.Login(username, password); err != nil {
// 			log.Err(err).Str("hostname", account.Hostname).Str("username", account.EncUser).
// 				Msg("Unable to log into to account.")
// 			continue
// 		}
// 		defer imapClient.Logout()
// 		log.Info().Str("username", account.EncUser).Msg("Successfully logged into account.")

// 		ap := &MailAccountManage{
// 			client:        imapClient,
// 			username:      username,
// 			accountConfig: &account,
// 		}
// 		mailboxes := ap.ListMailboxes()
// 		folders := lo.Map(mailboxes, func(m *imap.MailboxInfo, index int) string {
// 			return m.Name
// 		})
// 		log.Info().Strs("folders", folders).Msg("Found mailboxes in account")

// 		invalid := lo.Filter(account.Prune.Folders, func(f string, i int) bool {
// 			return !lo.Contains(folders, f) || !strings.Contains(strings.ToLower(f), "inbox")
// 		})
// 		log.Info().Msgf("Found %s invalid folders.", invalid)
// 		if len(invalid) != 0 {
// 			os.Exit(1)
// 		}

// 		mailboxes = lo.Filter(mailboxes, func(m *imap.MailboxInfo, index int) bool {
// 			return lo.Contains(folders, m.Name)
// 		})
// 		if err = ap.ProcessMailboxes(mailboxes); err != nil {
// 			log.Err(err).Str("username", account.EncUser).Msg("Failed to prune account with error.")
// 		}

// 	}

// 	return nil
// }

package main

import (
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/rs/zerolog/log"
)

type MailAccountManage struct {
	client        *client.Client
	mailboxes     []*imap.MailboxInfo
	username      string
	accountConfig *MailAccountConfig
}

func NewMailAccountManage(accountID int) (*MailAccountManage, error) {
	log.Info().Int("accountID", accountID).Msg("Fetching account from config with ID")
	c := getConfig()
	if len(*c.Mail.Accounts) < accountID {
		return nil, fmt.Errorf("%d is not a valid account ID", accountID)
	}

	account := (*c.Mail.Accounts)[accountID]

	var err error
	var username, password string

	decryptionKey := c.Encrypt.Secret
	decryptionIv := c.Encrypt.Iv

	log.Info().Int("accountID", accountID).
		Str("username", account.EncUser).Msg("Ingesting emails")

	username, err = Decrypt(account.EncUser, decryptionKey, decryptionIv)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt username with error %w", err)
	}

	password, err = Decrypt(account.EncPassword, decryptionKey, decryptionIv)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt password with error %w", err)
	}

	imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", account.Hostname, account.Port), nil)

	if err != nil {
		log.Err(err).Str("hostname", account.Hostname).Msg("Unable to connect to mail server.")
		return nil, fmt.Errorf("unable to connect to mail server %s with error %w", account.Hostname, err)
	}

	log.Info().Str("hostname", account.Hostname).Msg("Successfully connected to mail server")

	if err = imapClient.Login(username, password); err != nil {
		log.Err(err).Str("hostname", account.Hostname).Str("username", account.EncUser).
			Msg("Unable to log into to account.")
		return nil, fmt.Errorf("unable to login to the account")
	}
	log.Info().Str("username", account.EncUser).Msg("Successfully logged into account.")

	return &MailAccountManage{
		client:        imapClient,
		username:      username,
		mailboxes:     listMailboxes(account.EncUser, imapClient),
		accountConfig: &account,
	}, nil
}

func listMailboxes(username string, imapClient *client.Client) []*imap.MailboxInfo {
	log.Info().Str("enc_username", username).Msg("Fetching all mailboxes")
	mailboxes := make(chan *imap.MailboxInfo, 50)

	if err := imapClient.List("", "*", mailboxes); err != nil {
		log.Err(err).Str("enc_username", username).Msg("Unable to get folders from mailbox.")
		return []*imap.MailboxInfo{}
	}

	result := []*imap.MailboxInfo{}
	for m := range mailboxes {
		result = append(result, m)
	}
	return result
}

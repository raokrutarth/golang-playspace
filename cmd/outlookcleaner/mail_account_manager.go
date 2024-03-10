package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/raokrutarth/golang-playspace/pkg/logger"
)

type MailAccountConnection struct {
	client        *client.Client
	mailboxes     []imap.MailboxInfo
	username      string
	accountConfig MailAccountConfig
}

// NewMailAccountConnections get all the mail account credentials and init the imap clients
func NewMailAccountConnections(ctx context.Context) ([]MailAccountConnection, error) {
	log := logger.GetLogger()
	var err error

	log.Info("initializing mail account connections")
	c := getConfig()
	if len(c.Mail.Accounts) == 0 {
		return nil, fmt.Errorf("no mail accounts configured")
	}
	decryptionKey := c.Encrypt.Secret
	decryptionIv := c.Encrypt.Iv
	log.Info("identified credential decryption keys", "lenDecryptionKey", len(decryptionKey), "lenIV", len(decryptionIv))

	connections := []MailAccountConnection{}
	for _, account := range c.Mail.Accounts {
		var username, password string
		username, err = Decrypt(account.EncUser, decryptionKey, decryptionIv)
		if err != nil {
			return nil, fmt.Errorf("unable to decrypt username with error %w", err)
		}
		log = log.With("username", username)
		password, err = Decrypt(account.EncPassword, decryptionKey, decryptionIv)
		if err != nil {
			return nil, fmt.Errorf("unable to decrypt password with error %w", err)
		}
		log.Info("decrypted imap credentials", "username", username, "lenPwd", len(password))
		imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", account.Hostname, account.Port), nil)
		if err != nil {
			return nil, fmt.Errorf("unable to connect to mail server %s with error %w", account.Hostname, err)
		}
		if err = imapClient.Login(username, password); err != nil {
			return nil, fmt.Errorf("unable to login to host %s with error %w", account.Hostname, err)
		}
		log.Info("successfully logged into account", "username", username)

		folders, err := listMailboxes(logger.ContextWithLogger(ctx, log), account.EncUser, imapClient)
		if err != nil {
			return []MailAccountConnection{}, err
		}
		folderNames := []string{}
		for _, m := range folders {
			folderNames = append(folderNames, m.Name)
		}
		log.Info("listed folders in the mailbox", "folders", folderNames)

		// validate the configs
		for _, fn := range account.Ingest.Folders {
			if !slices.Contains(folderNames, fn) {
				return []MailAccountConnection{}, fmt.Errorf("folder %s does not exist for account %s", fn, username)
			}
		}
		for _, fn := range account.Prune.Folders {
			if !slices.Contains(folderNames, fn) {
				return []MailAccountConnection{}, fmt.Errorf("folder %s does not exist for account %s", fn, username)
			}
		}
		connections = append(connections, MailAccountConnection{
			client:        imapClient,
			username:      username,
			mailboxes:     folders,
			accountConfig: account,
		})
	}
	return connections, nil
}

func listMailboxes(ctx context.Context, username string, imapClient *client.Client) ([]imap.MailboxInfo, error) {
	mailboxes := []imap.MailboxInfo{}
	mailboxesSink := make(chan *imap.MailboxInfo, 50)
	if err := imapClient.List("", "*", mailboxesSink); err != nil {
		return mailboxes, fmt.Errorf("unable to list folders for user %s with error %w", username, err)
	}
	for m := range mailboxesSink {
		mailboxes = append(mailboxes, *m)
	}
	return mailboxes, nil
}

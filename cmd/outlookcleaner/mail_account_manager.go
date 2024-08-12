package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/raokrutarth/golang-playspace/pkg/logger"
)

type MailAccountConnection struct {
	client        *client.Client
	mailboxes     []imap.MailboxInfo
	username      string
	accountConfig MailAccountConfig
	startedAt     time.Time
}

// NewMailAccountConnections get all the mail account credentials and init the imap clients
func NewMailAccountConnections(ctx context.Context) ([]MailAccountConnection, error) {
	c := getConfig(ctx)
	if len(c.Mail.Accounts) == 0 {
		return nil, fmt.Errorf("no mail accounts configured")
	}

	l := logger.GetLoggerFromContext(ctx)
	l.Info("initializing mail account connections")

	connections := []MailAccountConnection{}
	for _, account := range c.Mail.Accounts {
		sl := l.With("encUsername", account.EncUser)
		imapClient, err := newIMAPClient(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize imap client with error: %w", err)
		}
		folders, listMailboxesErr := listMailboxes(account.EncUser, imapClient)
		if listMailboxesErr != nil {
			return nil, fmt.Errorf("unable to list mailboxes with error %w", listMailboxesErr)
		}
		folderNames := []string{}
		for _, m := range folders {
			folderNames = append(folderNames, m.Name)
		}
		sl.Info("listed folders in the mailbox", "folders", folderNames)

		// validate the configs
		for _, fn := range account.Ingest.Folders {
			if !slices.Contains(folderNames, fn) {
				return []MailAccountConnection{}, fmt.Errorf("folder %s does not exist for account %s to ingest", fn, account.EncUser)
			}
		}
		for _, fn := range account.Prune.Folders {
			if !slices.Contains(folderNames, fn) {
				return []MailAccountConnection{}, fmt.Errorf("folder %s does not exist for account %s to prune", fn, account.EncUser)
			}
		}
		connections = append(connections, MailAccountConnection{
			client:        imapClient,
			username:      account.EncUser,
			mailboxes:     folders,
			accountConfig: account,
			startedAt:     time.Now(),
		})
	}
	return connections, nil
}

// newIMAPClient: TODO convert this to a factory method since IMAP connections are flaky
// and the factory method checks for live-ness of the connection and creates a new connection
// without creating too many connections too fast (use the created at timestamp).
func newIMAPClient(ctx context.Context, account MailAccountConfig) (*client.Client, error) {
	l := logger.GetLoggerFromContext(ctx)
	c := getConfig(ctx)
	decryptionKey := c.Encrypt.Secret
	decryptionIv := c.Encrypt.Iv
	l.Info("identified credential decryption keys", "lenDecryptionKey", len(decryptionKey), "lenIv", len(decryptionIv))

	var err error
	var username, password string
	username, err = Decrypt(account.EncUser, decryptionKey, decryptionIv)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt username with error %w", err)
	}
	sl := l.With("username", username)
	password, err = Decrypt(account.EncPassword, decryptionKey, decryptionIv)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt password with error %w", err)
	}
	sl.Info("decrypted imap credentials", "lenPwd", len(password))
	imapClient, tlsErr := client.DialTLS(fmt.Sprintf("%s:%d", account.Hostname, account.Port), nil)
	if tlsErr != nil {
		return nil, fmt.Errorf("unable to connect to mail server %s with error %w", account.Hostname, tlsErr)
	}
	if err = imapClient.Login(username, password); err != nil {
		return nil, fmt.Errorf("unable to login to host %s with error %w", account.Hostname, err)
	}
	sl.Info("successfully logged into account")
	return imapClient, nil
}

func listMailboxes(username string, imapClient *client.Client) ([]imap.MailboxInfo, error) {
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

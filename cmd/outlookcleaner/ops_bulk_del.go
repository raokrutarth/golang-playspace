package main

import (
	"net/textproto"

	"github.com/emersion/go-imap"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func BulkDelete(from []string) {
	// var folderUnderUse string
	var ok bool
	var mailbox *Mailbox
	var mailboxInfo *imap.MailboxInfo
	var messageIDs []uint32
	// var seqSet *imap.SeqSet
	var err error
	// var accMgr *MailAccountManage

	accountMgr, err := NewMailAccountManage(0)
	if err != nil {
		log.Err(err).Msg("Failed to get account")
		return
	}
	defer accountMgr.client.Logout() // TODO find a better place for it

	// reread inbox messages to get stats
	folderUnderUse := "Inbox/INBOX"
	mailboxInfo, ok = lo.Find(accountMgr.mailboxes, func(m *imap.MailboxInfo) bool {
		return m.Name == folderUnderUse
	})
	if !ok {
		log.Error().Msg("Inbox folder not found")
		return
	}
	mailbox, err = NewMailbox(mailboxInfo, accountMgr.client)
	if err != nil {
		log.Err(err).Msg("Failed to initalize mailbox for folder")
		return
	}
	log.Info().Msgf("Selected a mailbox %s with %d messages", mailbox.status.Name, mailbox.status.Messages)
	criteria := &imap.SearchCriteria{
		Header: textproto.MIMEHeader{"From": {"technologyreview.com"}},
	}
	// criteria.Before = time.Now().Add(-3 * 365 * 24 * time.Hour)
	messageIDs, err = mailbox.client.Search(criteria)
	if err != nil {
		log.Err(err).Msg("Failed to search for fresh messages")
		return
	}
	log.Info().Msgf("Fetched %d message IDs from %s before %s", len(messageIDs), folderUnderUse, criteria.Before)

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(messageIDs...)
	log.Info().Msgf("Moving %d messages to %s", len(messageIDs),
		"Inbox/z-archive/to-delete")
	err = mailbox.client.Move(seqSet, "Inbox/z-archive/to-delete")
	if err != nil {
		log.Err(err).Msg("Failed to move messages to archive folder.")
		return
	}
}

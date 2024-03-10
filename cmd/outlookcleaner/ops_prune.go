package main

import (
	"context"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/ozgio/strutil"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func CustomPrune() {
	var folderUnderUse string
	var mailboxInfo imap.MailboxInfo
	var ok bool
	var mailbox *Mailbox
	var messageIDs []uint32
	var seqSet *imap.SeqSet
	var err error
	ctx := context.Background()

	accountMgrs, err := NewMailAccountConnections(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to get account")
		return
	}
	accountMgr := accountMgrs[0]
	defer accountMgr.client.Logout() // TODO find a better place for it

	// reread inbox messages to get stats
	folderUnderUse = "Inbox/z-archive"
	mailboxInfo, ok = lo.Find(accountMgr.mailboxes, func(m imap.MailboxInfo) bool {
		return m.Name == folderUnderUse
	})
	if !ok {
		log.Error().Msg("Inbox folder not found")
		return
	}
	mailbox, err = NewMailbox(&mailboxInfo, accountMgr.client)
	if err != nil {
		log.Err(err).Msg("Failed to initalize mailbox for folder")
		return
	}
	log.Info().Msgf("Selected a mailbox %s with %d messages", mailbox.status.Name, mailbox.status.Messages)
	criteria := &imap.SearchCriteria{}
	// criteria.Before = time.Now().Add(-3 * 365 * 24 * time.Hour)
	messageIDs, err = mailbox.client.Search(criteria)
	if err != nil {
		log.Err(err).Msg("Failed to search for fresh messages")
		return
	}
	log.Info().Msgf("Fetched %d message IDs from %s before %s", len(messageIDs), folderUnderUse, criteria.Before)

	done := make(chan error, 1)
	messages := make(chan *imap.Message, 10)
	go func() {
		seqSet = new(imap.SeqSet)
		seqSet.AddNum(messageIDs...)
		items := []imap.FetchItem{
			imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822Header,
			imap.FetchRFC822Size, imap.FetchUid, imap.FetchBodyStructure,
		}
		done <- mailbox.client.Fetch(seqSet, items, messages)
	}()

	receiptMessageIDs, flaggedMessageIDs, attachmentMessageIDs, staleMessageIDs := []uint32{}, []uint32{}, []uint32{}, []uint32{}
	for msg := range messages {
		log.Debug().Msgf("Got email for mailbox %s: %s", mailbox.info.Name, msg.Envelope.Subject)
		log.Debug().Msgf(
			"flags: %v, senders: %+v, from: %+v ID: %v, date: %v",
			msg.Flags, msg.Envelope.Sender, msg.Envelope.From, msg.Envelope.MessageId, msg.Envelope.Date,
		)
		_, isFlagged := lo.Find(msg.Flags, func(f string) bool {
			return f == imap.FlaggedFlag || f == imap.ImportantFlag
		})
		if isFlagged {
			flaggedMessageIDs = append(flaggedMessageIDs, msg.SeqNum)
		}

		isReceipt := lo.ContainsBy(strutil.Words(msg.Envelope.Subject), func(w string) bool {
			return lo.Contains([]string{
				"refund", "shipped", "receipt",
				"order", "confirm", "boarding", "delivered",
				"reservation",
			}, strings.ToLower(w))
		})
		if isReceipt {
			receiptMessageIDs = append(receiptMessageIDs, msg.SeqNum)
		}

		var attachments []string
		msg.BodyStructure.Walk(func(path []int, part *imap.BodyStructure) bool {
			if !strings.EqualFold(part.Disposition, "attachment") {
				return true
			}

			filename, _ := part.Filename()
			log.Debug().Msgf("Message %s has attachment of type: %s, filename: %s",
				msg.Envelope.MessageId,
				strings.ToLower(part.MIMEType+"/"+part.MIMESubType),
				filename,
			)

			attachments = append(attachments, filename)
			return true
		})
		if len(attachments) != 0 {
			attachmentMessageIDs = append(attachmentMessageIDs, msg.SeqNum)
		}

		staleMessageIDs = append(staleMessageIDs, msg.SeqNum)
	}
	_ = staleMessageIDs

	ops := []struct {
		ids        []uint32
		destFolder string
	}{
		// {ids: staleMessageIDs, destFolder: "Inbox/z-archive"},
		{ids: flaggedMessageIDs, destFolder: "Inbox/z-archive/flagged"},
		{ids: attachmentMessageIDs, destFolder: "Inbox/z-archive/has-attachment"},
		{ids: receiptMessageIDs, destFolder: "Inbox/z-archive/receipt"},
	}

	for _, op := range ops {
		if len(op.ids) == 0 {
			log.Info().Msgf("No messages to move to %s", op.destFolder)
			continue
		}
		seqSet = new(imap.SeqSet)
		seqSet.AddNum(op.ids...)
		log.Info().Msgf("Moving %d messages to %s", len(op.ids), op.destFolder)
		err = mailbox.client.Move(seqSet, op.destFolder)
		if err != nil {
			log.Err(err).Msg("Failed to move messages to archive folder.")
			return
		}
	}

	log.Info().Msg("Finished pruning all messages")
	if err = <-done; err != nil {
		log.Err(err).Msg("Failed to list messages for states with error")
	}
}

func CustomPrune2() {
	// TODO: fix everything moved to z-archive
	ctx := context.Background()

	accountMgrs, err := NewMailAccountConnections(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to get account")
		return
	}
	accountMgr := accountMgrs[0]
	defer accountMgr.client.Logout() // TODO find a better place for it

	var folderUnderUse string
	var mailboxInfo imap.MailboxInfo
	var ok bool
	var mailbox *Mailbox
	var messageIDs []uint32

	// Move all incorrectly moved messages back to inbox
	folderUnderUse = "Inbox/z-archive"
	mailboxInfo, ok = lo.Find(accountMgr.mailboxes, func(m imap.MailboxInfo) bool {
		return m.Name == folderUnderUse
	})
	if !ok {
		log.Error().Msg("Inbox folder not found")
		return
	}
	mailbox, err = NewMailbox(&mailboxInfo, accountMgr.client)
	if err != nil {
		log.Err(err).Msg("Failed to initalize mailbox for folder")
		return
	}
	log.Info().Msgf("Selected a mailbox %s with %d messages", mailbox.status.Name, mailbox.status.Messages)
	log.Info().Msgf("Fetching  in folder %s", folderUnderUse)
	// criteria := imap.NewSearchCriteria()
	// criteria.Since = time.Now().Add(-5 * 365 * 24 * time.Hour) // delivered at least 5 years ago
	messageIDs, err = mailbox.client.Search(&imap.SearchCriteria{})
	if err != nil {
		log.Err(err).Msg("Failed to search for fresh messages")
		return
	}
	log.Info().Msgf("Found %d messages from %s", len(messageIDs), folderUnderUse)
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(messageIDs...)
	log.Info().Msg("Moving messages back to Inbox")
	err = mailbox.client.Move(seqSet, "INBOX")
	if err != nil {
		log.Err(err).Msg("Failed to move messages back to inbox")
		return
	}
}

func CustomPrune1() {
	// TODO: fix everything moved to z-archive
	ctx := context.Background()
	accountMgrs, err := NewMailAccountConnections(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to get account")
		return
	}
	accountMgr := accountMgrs[0]
	defer accountMgr.client.Logout() // TODO find a better place for it
	if err != nil {
		log.Err(err).Msg("Failed to get account")
		return
	}

	folderUnderUse := "INBOX"

	mailboxInfo, ok := lo.Find(accountMgr.mailboxes, func(m imap.MailboxInfo) bool {
		return m.Name == folderUnderUse
	})
	if !ok {
		log.Error().Msg("Inbox folder not found")
		return
	}

	mailbox, err := NewMailbox(&mailboxInfo, accountMgr.client)
	if err != nil {
		log.Err(err).Msg("Failed to initalize mailbox for folder")
		return
	}
	log.Info().Str("mailbox_name", mailbox.status.Name).Uint32("num_messages", mailbox.status.Messages).
		Msgf("Selected a mailbox folder")

	log.Info().Msg("Fetching messages marked deleted")
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.DeletedFlag}
	ids, err := mailbox.client.Search(criteria)
	if err != nil {
		log.Err(err).Msg("Failed to get messages marked for delete")
		return
	}
	log.Info().Msgf("Fetched %d messages marked for delete", len(ids))

	log.Info().Msg("Removing deletion flag from messages")
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(ids...)
	mailbox.client.Store(
		seqSet,
		imap.FormatFlagsOp(imap.RemoveFlags, true),
		[]interface{}{imap.DeletedFlag},
		nil,
	)

	log.Info().Msg("Moving messages to review folder")
	err = mailbox.client.Move(seqSet, "Inbox/z-archive")
	if err != nil {
		log.Err(err).Msg("Failed to move messages")
		return
	}
}

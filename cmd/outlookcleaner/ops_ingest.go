package main

import (
	"strings"
	"unicode"

	"github.com/emersion/go-imap"
	"github.com/ozgio/strutil"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// func Ingest() {
// 	var folderUnderUse string
// 	var ok bool
// 	var mailbox *Mailbox
// 	var messageIDs []uint32
// 	var seqSet *imap.SeqSet
// 	var err error
// 	// var accMgr *MailAccountManage

// 	// accountMgr, err := NewMailAccountManage(0)
// 	// if err != nil {
// 	// 	log.Err(err).Msg("Failed to get account")
// 	// 	return
// 	// }
// 	// defer accountMgr.client.Logout() // TODO find a better place for it

// 	// reread inbox messages to get stats
// 	foldersToIngest := []string{
// 		// "INBOX",
// 		// "Inbox/home",
// 		"Inbox/events",
// 		// "Inbox/automation",
// 		// "Inbox/project N",
// 		// "Inbox/project T",
// 		// "Inbox/personal/legal",
// 		// "Inbox/personal",
// 		// "Inbox/jobs",
// 	}

// 	for _, mailboxInfo := range accountMgr.mailboxes {
// 		folderUnderUse = mailboxInfo.Name

// 		if !lo.Contains(foldersToIngest, folderUnderUse) {
// 			log.Info().Msgf("Skipping folder %s at ingest time.", folderUnderUse)
// 			continue
// 		}
// 		mailboxInfo, ok = lo.Find(accountMgr.mailboxes, func(m *imap.MailboxInfo) bool {
// 			return m.Name == folderUnderUse
// 		})
// 		if !ok {
// 			log.Error().Msg("Inbox folder not found")
// 			return
// 		}

// 		accMgr, err = NewMailAccountManage(0)
// 		if err != nil {
// 			log.Err(err).Msg("Failed to get fresh account connection.")
// 			return
// 		}
// 		mailbox, err = NewMailbox(mailboxInfo, accMgr.client)
// 		if err != nil {
// 			log.Err(err).Msg("Failed to initalize mailbox for folder")
// 			return
// 		}
// 		log.Info().Msgf("Selected a mailbox %s with %d messages", mailbox.status.Name, mailbox.status.Messages)
// 		messageIDs, err = mailbox.client.Search(&imap.SearchCriteria{})
// 		if err != nil {
// 			log.Err(err).Msg("Failed to search for fresh messages")
// 			return
// 		}
// 		log.Info().Msgf("Fetched %d message IDs from %s", len(messageIDs), folderUnderUse)

// 		done := make(chan error, 1)
// 		messages := make(chan *imap.Message, 10)
// 		var section imap.BodySectionName
// 		go func() {
// 			seqSet = new(imap.SeqSet)
// 			seqSet.AddNum(messageIDs...)
// 			items := []imap.FetchItem{
// 				imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822Header,
// 				imap.FetchRFC822Size, imap.FetchUid, imap.FetchBodyStructure,
// 				section.FetchItem(),
// 			}
// 			done <- mailbox.client.Fetch(seqSet, items, messages)
// 		}()

// 		log.Info().Msgf("Processing %d messages in %s", len(messageIDs), folderUnderUse)
// 		for msg := range messages {
// 			// see header content for sender search.

// 			log.Debug().Msgf("Got email for mailbox %s: %s", folderUnderUse, msg.Envelope.Subject)
// 			log.Debug().Msgf(
// 				"flags: %v, senders: %+v, from: %+v, reply-to: %+v, ID: %v, date: %v",
// 				msg.Flags, msg.Envelope.Sender, msg.Envelope.From, msg.Envelope.ReplyTo, msg.Envelope.MessageId, msg.Envelope.Date,
// 			)

// 			var dbRecord *Message
// 			dbRecord, err = messageToDBRecord(msg, section)
// 			if err != nil {
// 				log.Err(err).Msg("failed to parse message with error")
// 				continue
// 			}
// 			dbRecord.MailBoxFolder = folderUnderUse

// 			dbWriteResult := GormDB.Clauses(clause.OnConflict{
// 				Columns:   []clause.Column{{Name: "message_id"}},
// 				UpdateAll: true,
// 			}).Create(dbRecord) // pass pointer of data to Create

// 			log.Debug().Msgf(
// 				"Wrote record to database with numUpdates: %d",
// 				dbWriteResult.RowsAffected, // returns inserted records count)
// 			)
// 			if dbWriteResult.Error != nil {
// 				log.Err(err).Msgf("Failed to write to DB with error for message %s", dbRecord.MessageID)
// 				return
// 			}
// 		}

// 		log.Info().Msgf("Finished processing all messages from folder %s", folderUnderUse)
// 		if err = <-done; err != nil {
// 			log.Err(err).Msg("Failed to list messages for states with error")
// 		}

// 		if err = accMgr.client.Logout(); err != nil {
// 			log.Err(err).Msgf("Failed to log out of folder %s with error.", folderUnderUse)
// 		}
// 	}
// }

func messageToDBRecord(msg *imap.Message, section imap.BodySectionName) (*Message, error) {
	if msg == nil {
		log.Error().Msg("Server didn't returned message")
	}
	dbRecord := &Message{}

	var sender *imap.Address

	switch sn, fn, rn := len(msg.Envelope.Sender), len(msg.Envelope.From), len(msg.Envelope.ReplyTo); {
	case rn > 1:
		log.Warn().Msgf("ignoring reply-to addresses %+v", msg.Envelope.ReplyTo[1:])
		fallthrough
	case rn == 1:
		sender = msg.Envelope.ReplyTo[0]
	case sn > 1:
		log.Warn().Msgf("ignoring senders %+v", msg.Envelope.Sender[1:])
		fallthrough
	case sn == 1:
		sender = msg.Envelope.Sender[0]
	case fn > 1:
		log.Warn().Msgf("ignoring from addresses %+v", msg.Envelope.From[1:])
		fallthrough
	case fn == 1:
		sender = msg.Envelope.From[0]
	default:
		log.Error().Msgf("Message %s has no sender addresses", msg.Envelope.MessageId)
		sender = &imap.Address{MailboxName: "unknown", HostName: "unknown"}
	}
	dbRecord.MessageID = msg.Envelope.MessageId
	dbRecord.From = sender.Address()
	dbRecord.FromName = sender.PersonalName
	dbRecord.ReceivedAt = msg.Envelope.Date
	dbRecord.Subject = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, msg.Envelope.Subject)
	dbRecord.SizeBytes = msg.Size

	_, isSeen := lo.Find(msg.Flags, func(f string) bool {
		return f == imap.SeenFlag
	})
	if isSeen {
		dbRecord.IsSeen = true
	}

	_, isFlagged := lo.Find(msg.Flags, func(f string) bool {
		return f == imap.FlaggedFlag || f == imap.ImportantFlag
	})
	if isFlagged {
		dbRecord.IsFlagged = true
	}

	isReceipt := lo.ContainsBy(strutil.Words(msg.Envelope.Subject), func(w string) bool {
		return lo.Contains([]string{
			"refund", "shipped", "receipt", "confirmation",
			"order", "confirm", "boarding", "delivered",
			"reservation",
		}, strings.ToLower(w))
	})
	if isReceipt {
		dbRecord.IsReceipt = true
	}

	// get attachment names
	var attachments []string
	msg.BodyStructure.Walk(func(path []int, part *imap.BodyStructure) bool {
		if !strings.EqualFold(part.Disposition, "attachment") {
			return true
		}

		filename, _ := part.Filename()
		log.Info().Msgf("Message %+v has attachment of type %s, filename: %s",
			path,
			strings.ToLower(part.MIMEType+"/"+part.MIMESubType),
			filename,
		)

		attachments = append(attachments, filename)
		return true
	})
	attachmentNames := strings.Join(attachments, "#")
	dbRecord.AttachmentNames = attachmentNames

	// bodyAttrs := map[string]string{}
	// r := msg.GetBody(&section)
	// if r == nil {
	// 	return nil, errors.New("received empty body")
	// }
	// // Create a new mail reader
	// mr, err := mail.CreateReader(r)
	// if err != nil {
	// 	log.Err(err).Msg("failed to create message reader with error")
	// }

	// // Process each message's part
	// i := 0
	// var b []byte
	// var p *mail.Part
	// for {
	// 	p, err = mr.NextPart()
	// 	if errors.Is(err, io.EOF) {
	// 		break
	// 	} else if err != nil {
	// 		log.Err(err).Msg("failed to read message next part")
	// 	}

	// 	switch p.Header.(type) {
	// 	case *mail.InlineHeader:
	// 		// This is the message's text (can be plain-text or HTML)
	// 		b, err = ioutil.ReadAll(p.Body)
	// 		if err != nil {
	// 			log.Err(err).Msgf("Failed to read email body with error %s", err)
	// 		}
	// 		bodyAttrs[fmt.Sprintf("body_%d", i)] = string(b)
	// 	}
	// 	i++
	// }

	// if len(bodyAttrs) != 0 {
	// 	dbRecord.Attributes, err = json.Marshal(bodyAttrs)
	// 	if err != nil {
	// 		log.Err(err).Msgf("failed to encode body contents into attributes with error.")
	// 	}
	// }

	return dbRecord, nil
}

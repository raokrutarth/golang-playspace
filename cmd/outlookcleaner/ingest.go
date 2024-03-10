package main

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/ozgio/strutil"
	"github.com/raokrutarth/golang-playspace/pkg/logger"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"gorm.io/gorm/clause"
)

// https://github.com/search?l=Go&p=3&q=%22emersion%2Fgo-imap%22&type=Code
// attachments: https://github.com/wireless-road/warehouse/blob/06329db2d57601e583584629d6f146f54e0bbcfa/src/api/email/email.go#L74
// date filter https://github.com/fronge/studyGo/blob/2ed01b658a5afee65d78d74843767b3e52348d77/emailT/main.go#L48
// new messages: https://github.com/budenny/mail2telegram/blob/4b77de2b54ba482eccb907d40fa329f4757d8843/mail/client.go#L44
// attachments: https://github.com/axllent/imap-scrub/blob/1183564f59a3f441e2ffccbb891e7a62d2ea2f53/lib/parser.go#L25
// write officedocument.wordprocessingml to file https://github.com/triano4/golang/blob/a3e81cefd926fd87e832942cd3c57225defd2d6c/code/Client.go#L82
// move folders: https://github.com/thedustin/go-email-curator/blob/54c33f2d542d4c20e8a72fc03a0d88fc9d118253/action/move.go
// https://github.com/donomii/shonkr/blob/6261545c6d47c623fe6043fec2130f4129018e1f/v3/getmail.go

type accountIngestor struct {
	imapClient *client.Client
}

func (ingestor *accountIngestor) getUnreadMessages() []uint32 {
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}
	uids, err := ingestor.imapClient.Search(criteria)
	if err != nil {
		log.Err(err).Msg("Unable to search for unread messages")
	}
	return uids
}

func (ingestor *accountIngestor) fetchMessages(mailbox *imap.MailboxStatus, seqNums []uint32) []map[string]string {

	seqset := new(imap.SeqSet)
	seqset.AddNum(seqNums...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- ingestor.imapClient.Fetch(
			seqset, []imap.FetchItem{
				// imap.FetchEnvelope, imap.FetchRFC822Text, imap.FetchBody,
				// imap.FetchUid, imap.FetchRFC822Text,
				imap.FetchFull,
			}, messages,
		)
	}()

	out := make([]map[string]string, 0)
	for msg := range messages {
		out = append(out, ParseMessage(msg))
	}

	if err := <-done; err != nil {
		log.Err(err).Msg("Failed to finish fetching messages.")
	}
	return out
}

func (ingestor *accountIngestor) fetchNewMessages(folders []string) (result []*imap.Message, err error) {

	for _, folder := range folders {
		mbox, err := ingestor.imapClient.Select(folder, false)
		if err != nil {
			return result, err
		}

		from := uint32(1)
		to := mbox.Messages

		if mbox.Messages > 2 {
			// We're using unsigned integers here, only subtract if the result is > 0
			from = mbox.Messages - 2
		}
		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		messages := make(chan *imap.Message, 100)

		done := make(chan error, 1)
		// go func() {
		// 	done <- ingestor.imapClient.Fetch(seqset, []imap.FetchItem{imap.FetchBody}, messages)
		// }()

		// for msg := range messages {
		// 	result = append(result, msg)
		// }

		go func() {
			done <- ingestor.imapClient.Fetch(
				seqset,
				[]imap.FetchItem{
					imap.FetchUid, imap.FetchRFC822Text, imap.FetchBody,
					imap.FetchEnvelope, imap.FetchBodyStructure, imap.FetchFlags,
				},
				messages,
			)
		}()
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{"\\Seen"}
		_, err = ingestor.imapClient.Search(criteria)

		// log.Println("Last 4 messages:")
		var out [][]string
		for msg := range messages {
			data := fmt.Sprintf("%+v, %+v", msg.Envelope, msg.BodyStructure)
			for _, v := range msg.Body {
				// fmt.Println("Body: '", k, "'", v)
				data = fmt.Sprintf("%v", v)
			}
			// fmt.Println(data)
			out = append(out, []string{msg.Envelope.Subject, data})
		}

		if err := <-done; err != nil {
			log.Err(err)
		}
		return result, err
	}

	return result, err
}

func Ingest(ctx context.Context, connections []MailAccountConnection) error {
	var err error
	for _, conn := range connections {
		for _, mInfo := range conn.mailboxes {
			err = ingestMailbox(ctx, conn, mInfo)
			if err != nil {
				return fmt.Errorf("unable to ingest mailbox with error %w", err)
			}
		}
	}
	return nil
}

func ingestMailbox(ctx context.Context, conn MailAccountConnection, mailboxInfo imap.MailboxInfo) error {
	log := logger.GetLoggerFromContext(ctx).With("folderName", mailboxInfo.Name)

	folderUnderUse := mailboxInfo.Name
	status, err := conn.client.Select(mailboxInfo.Name, true)
	if err != nil {
		return fmt.Errorf("unable to select folder %s with error %w", mailboxInfo.Name, err)
	}
	log.Info("selected a mailbox", "numMessages", status.Messages, "numUnread", status.Unseen, "recent", status.Recent)

	var messageIDs []uint32
	var seqSet *imap.SeqSet
	messageIDs, err = conn.client.Search(&imap.SearchCriteria{})
	if err != nil {
		return fmt.Errorf("failed to search folder %s wit error %w", mailboxInfo.Name, err)
	}

	done := make(chan error, 1)
	messages := make(chan *imap.Message, 10)
	// var section imap.BodySectionName
	go func() {
		seqSet = new(imap.SeqSet)
		seqSet.AddNum(messageIDs...)
		items := []imap.FetchItem{
			imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822Header,
			imap.FetchRFC822Size, imap.FetchUid,
			// imap.FetchBodyStructure,
			// section.FetchItem(),
		}
		done <- conn.client.Fetch(seqSet, items, messages)
	}()

	log.Info("processing messages", "numMessageIDs", len(messageIDs))
	for msg := range messages {
		// see header content for sender search.

		log.Debug().Msgf("Got email for mailbox %s: %s", folderUnderUse, msg.Envelope.Subject)
		log.Debug().Msgf(
			"flags: %v, senders: %+v, from: %+v, reply-to: %+v, ID: %v, date: %v",
			msg.Flags, msg.Envelope.Sender, msg.Envelope.From, msg.Envelope.ReplyTo, msg.Envelope.MessageId, msg.Envelope.Date,
		)

		var dbRecord *Message
		dbRecord, err = messageToDBRecord(msg, section)
		if err != nil {
			log.Err(err).Msg("failed to parse message with error")
			continue
		}
		dbRecord.MailBoxFolder = folderUnderUse

		dbWriteResult := GormDB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}},
			UpdateAll: true,
		}).Create(dbRecord) // pass pointer of data to Create

		log.Debug().Msgf(
			"Wrote record to database with numUpdates: %d",
			dbWriteResult.RowsAffected, // returns inserted records count)
		)
		if dbWriteResult.Error != nil {
			log.Err(err).Msgf("Failed to write to DB with error for message %s", dbRecord.MessageID)
			return
		}
	}

	log.Info().Msgf("Finished processing all messages from folder %s", folderUnderUse)
	if err = <-done; err != nil {
		log.Err(err).Msg("Failed to list messages for states with error")
	}

	if err = accMgr.client.Logout(); err != nil {
		log.Err(err).Msgf("Failed to log out of folder %s with error.", folderUnderUse)
	}
}

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

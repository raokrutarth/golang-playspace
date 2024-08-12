package main

import (
	"context"
	"fmt"
	"strings"
	"time"
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
	for _, conn := range connections { // TODO iterate over accounts instead of connections since connects are flaky and need to be created anew each time
		for _, mInfo := range conn.mailboxes {
			// FIXME: reset connection due to unknown timeout
			// t="2024-08-12 03:09:51" level=ERROR s=main.go:56 msg="failed to ingest" cmd=ingest error="unable to ingest mailbox with error unable to select folder Inbox/personal/cashtrac with error User is authenticated but not connected."
			conn.client, err = newIMAPClient(ctx, conn.accountConfig)
			if err != nil {
				return fmt.Errorf("unable to create new imap client with error %w", err)
			}
			err = ingestMailbox(ctx, conn, mInfo)
			if err != nil {
				return fmt.Errorf("unable to ingest mailbox with error %w", err)
			}
			conn.client.Close()
			time.Sleep(time.Second * 30)
		}
	}
	return nil
}

func ingestMailbox(ctx context.Context, conn MailAccountConnection, mailboxInfo imap.MailboxInfo) error {
	folderUnderUse := mailboxInfo.Name
	l := logger.GetLoggerFromContext(ctx).With("folderName", folderUnderUse)
	status, err := conn.client.Select(mailboxInfo.Name, true)
	if err != nil {
		return fmt.Errorf("unable to select folder %s with error %w", mailboxInfo.Name, err)
	}
	l.Info("selected a mailbox", "numMessages", status.Messages, "numUnread", status.Unseen, "recent", status.Recent)
	if status.Messages == 0 {
		l.Info("no messages in folder")
		return nil
	}

	var messageIDs []uint32
	var seqSet *imap.SeqSet
	messageIDs, err = conn.client.Search(&imap.SearchCriteria{
		// Text: []string{"robinhood"}, // used to debug when risking deleting messages
	})
	if err != nil {
		return fmt.Errorf("failed to search folder %s with error: %w", mailboxInfo.Name, err)
	}
	if len(messageIDs) == 0 {
		l.Warn("no messages IDs in folder for search")
		return nil
	}

	done := make(chan error, 1)
	messages := make(chan *imap.Message, 10)
	// var section imap.BodySectionName // FIXME: fetching the body marks the message as read
	go func() {
		seqSet = new(imap.SeqSet)
		seqSet.AddNum(messageIDs...)
		items := []imap.FetchItem{
			imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822Header,
			imap.FetchRFC822Size, imap.FetchUid, imap.FetchInternalDate,
			imap.FetchRFC822, imap.FetchRFC822Header, imap.FetchBodyStructure,

			// FIXME: fetching the body marks the message as read
			// imap.FetchBodyStructure,
			// section.FetchItem(),
		}
		done <- conn.client.Fetch(seqSet, items, messages)
	}()

	l.Info("processing messages", "numMessageIds", len(messageIDs))
	for msg := range messages {
		sl := l.With("messageID", msg.Uid, "subject", msg.Envelope.Subject)
		sl.Debug(
			"got email from mailbox",
			"subject", msg.Envelope.Subject, "flags", msg.Flags, "sender", msg.Envelope.Sender,
			"from", msg.Envelope.From, "replyTo", msg.Envelope.ReplyTo, "ID", msg.Envelope.MessageId,
			"date", msg.Envelope.Date, "sizeBytes", msg.Size, "uid", msg.Uid,
		)

		var dbRecord *Message
		dbRecord, err = messageToDBRecord(logger.ContextWithLogger(ctx, sl), msg)
		if err != nil {
			sl.Error("failed to parse message with error", "error", err)
			continue
		}
		dbRecord.MailBoxFolder = folderUnderUse
		dbWriteResult := GormDB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}},
			UpdateAll: true,
		}).Create(dbRecord) // pass pointer of data to Create

		sl.Debug(
			"wrote record to database",
			"numUpdates", dbWriteResult.RowsAffected, // returns inserted records count)
		)
		if dbWriteResult.Error != nil {
			return fmt.Errorf("failed to write to DB with error for message: %w", dbWriteResult.Error)
		}
		if dbWriteResult.RowsAffected == 0 {
			sl.Error("no record added to DB", "subject", msg.Envelope.Subject)
		}
	}
	if err = <-done; err != nil {
		return fmt.Errorf("failed to fetch all messages with error: %w", err)
	}
	l.Info("finished processing messages", "numMessageIds", len(messageIDs))
	return nil
}

func messageToDBRecord(ctx context.Context, msg *imap.Message) (*Message, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}
	l := logger.GetLoggerFromContext(ctx)

	var sender *imap.Address
	switch sn, fn, rn := len(msg.Envelope.Sender), len(msg.Envelope.From), len(msg.Envelope.ReplyTo); {
	case rn > 1:
		l.Debug("ignoring reply-to addresses", "otherReplyTo", msg.Envelope.ReplyTo[1:])
		fallthrough
	case rn == 1:
		sender = msg.Envelope.ReplyTo[0]
	case sn > 1:
		l.Debug("ignoring senders", "otherSenders", msg.Envelope.Sender[1:])
		fallthrough
	case sn == 1:
		sender = msg.Envelope.Sender[0]
	case fn > 1:
		l.Debug("ignoring from addresses", "otherFrom", msg.Envelope.From[1:])
		fallthrough
	case fn == 1:
		sender = msg.Envelope.From[0]
	default:
		l.Error("message has no sender addresses", "messageID", msg.Envelope.MessageId)
		sender = &imap.Address{MailboxName: "unknown", HostName: "unknown"}
	}
	dbRecord := &Message{}
	dbRecord.MessageID = msg.Envelope.MessageId
	dbRecord.From = sender.Address()
	dbRecord.FromName = sender.PersonalName
	if len(msg.Envelope.To) > 0 {
		dbRecord.To = msg.Envelope.To[0].Address()
	}
	dbRecord.ReceivedAt = msg.Envelope.Date
	dbRecord.Subject = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, msg.Envelope.Subject)
	dbRecord.SizeBytes = msg.Size
	dbRecord.UID = msg.Uid
	dbRecord.SeqNum = msg.SeqNum

	if _, isSeen := lo.Find(msg.Flags, func(f string) bool {
		return f == imap.SeenFlag
	}); isSeen {
		dbRecord.IsSeen = true
	}

	if _, isFlagged := lo.Find(msg.Flags, func(f string) bool {
		return f == imap.FlaggedFlag || f == imap.ImportantFlag
	}); isFlagged {
		dbRecord.IsFlagged = true
	}

	if isReceipt := lo.ContainsBy(strutil.Words(msg.Envelope.Subject), func(w string) bool {
		return lo.Contains([]string{
			"refund", "shipped", "receipt", "confirmation",
			"order", "confirm", "boarding", "delivered",
			"reservation",
		}, strings.ToLower(w))
	}); isReceipt {
		dbRecord.IsReceipt = true
	}

	// get attachment names
	var attachments []string
	msg.BodyStructure.Walk(func(path []int, part *imap.BodyStructure) bool {
		if !strings.EqualFold(part.Disposition, "attachment") {
			return true
		}
		filename, _ := part.Filename()
		l.Debug(
			"found attachment",
			"filename", filename, "path", path, "type", strings.ToLower(part.MIMEType+"/"+part.MIMESubType),
		)
		attachments = append(attachments, filename)
		return true
	})
	attachmentNames := strings.Join(attachments, "#")
	dbRecord.AttachmentNames = attachmentNames

	// FIXME: trying to read the body implicitly marks the message as seen
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

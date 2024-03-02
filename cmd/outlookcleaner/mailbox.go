package main

import (
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/rs/zerolog/log"
)

type Mailbox struct {
	status *imap.MailboxStatus
	info   *imap.MailboxInfo
	client *client.Client
}

// CAUTION: this selects the mailbox folder on the client globally.
func NewMailbox(info *imap.MailboxInfo, imapClient *client.Client) (*Mailbox, error) {
	// TODO get DB row
	mailboxStatus, err := imapClient.Select(info.Name, false)
	if err != nil {
		log.Err(err).Str("folder", info.Name).Msgf("Unable to select folder")
		return nil, fmt.Errorf("unable to select inbox %s", info.Name)
	}

	return &Mailbox{
		status: mailboxStatus,
		info:   info,
		client: imapClient,
	}, nil
}

func (mbox *Mailbox) getLastScrapeTime() time.Time {
	return time.Now().Add(-5 * 365 * 24 * time.Hour)
}

// store store status without login
func (mbox *Mailbox) storeStatus(mailBox string, mID uint32, isAdd bool, flags []interface{}) error {

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(mID)

	var opt imap.FlagsOp
	if isAdd {
		opt = imap.AddFlags
	} else {
		opt = imap.RemoveFlags
	}

	item := imap.FormatFlagsOp(opt, true)

	return mbox.client.Store(seqSet, item, flags, nil)
}

// SetRead set read status
// func (mbox *Mailbox) SetRead(isRead bool) error {
// 	return mbox.client.Store(m.Box, m.ID, isRead, []interface{}{imap.SeenFlag})
// }

// DeleteMail delete one mail
func (mbox *Mailbox) DeleteMail(mailBox string, mID uint32) error {
	// First mark the message as deleted
	if err := mbox.storeStatus(mailBox, mID, true, []interface{}{imap.DeletedFlag}); err != nil {
		return err
	}

	// Then delete it
	deletedCh := make(chan uint32)
	err := mbox.client.Expunge(deletedCh)
	go func() {
		defer close(deletedCh)
		for mid := range deletedCh {
			log.Info().Msgf("Deleted message with ID %d", mid)
		}
	}()
	return err
}

func (mbox *Mailbox) GetUnReadMailIDs(mailBox string) ([]uint32, error) {
	if len(mailBox) == 0 {
		mailBox = "INBOX"
	}

	// Select mail box
	_, err := mbox.client.Select(mailBox, false)
	if err != nil {
		return nil, err
	}

	// Set search criteria
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := mbox.client.Search(criteria)
	if err != nil {
		return nil, err
	}

	return ids, err
}

// func (mbox *Mailbox) PruneStaleMessages(config configtypes.MailboxActionConfig) error {
// 	lastScrapeTime := mbox.getLastScrapeTime()

// 	log.Info().Time("scraped_at", lastScrapeTime).Str("mailbox_name", mbox.status.Name).Msg("Starting stalenes prune")
// 	uids := mbox.getMessageIdsInWindow(time.Time{}, lastScrapeTime)
// 	if len(uids) == 0 {
// 		log.Warn().Str("mailbox_name", mbox.status.Name).Msg("No new messages.")
// 		return nil
// 	}

// 	if configtypes.GlobalConfig.IsInDev() {
// 		uids = lo.Samples(uids, 100)
// 		log.Debug().Str("mailbox_name", mbox.status.Name).Uints32("message_ids", uids).Msg("Running in debug mode. Sampeling message IDs")
// 	}

// 	messages := make(chan *imap.Message, 50)
// 	seqset := new(imap.SeqSet)
// 	seqset.AddNum(uids...)

// 	var section imap.BodySectionName
// 	items := []imap.FetchItem{section.FetchItem()}
// 	go func() {
// 		if err := mbox.client.Fetch(seqset, items, messages); err != nil {
// 			log.Err(err).Str("mailbox_name", mbox.status.Name).Msg("failed to fetch messages with error.")
// 		}
// 	}()

// 	i := 0
// 	toDel := []uint32{}

// 	for msg := range messages {
// 		parsed := ParseMessageV2(msg, section, mbox)
// 		log.Info().Msgf("parsed email for mailbox %s: %s", mbox.info.Name, strutil.Summary(parsed["body_0"], 10, "..."))
// 		file, _ := json.MarshalIndent(parsed, "", " ")
// 		os.Exit(1)

// 		filePrefix := "./data/" + strutil.ToSnakeCase(strutil.Slugify(mbox.info.Name))
// 		if mid, ok := parsed["message_id"]; ok {
// 			filePrefix += "-" + strutil.Slugify(mid)
// 		} else {
// 			filePrefix += fmt.Sprintf("-%d", i)
// 		}
// 		_ = ioutil.WriteFile(filePrefix+"-parsed"+".json", file, 0600)
// 		for k := range parsed {
// 			if strings.Contains(k, "body") {
// 				_ = ioutil.WriteFile(filePrefix+".html", []byte(parsed[k]), 0600)
// 			}
// 		}
// 		i++

// 		var ok bool
// 		if _, ok = parsed["attachment_0"]; ok {
// 			continue
// 		}
// 		receivedAt, ok := parsed["date"]
// 		if !ok {
// 			log.Error().Msgf("Missing date for message with subject %s in mailbox %s", parsed["subject"], mbox.info.Name)
// 		}
// 		log.Info().Msgf("Adding message subject %s and id %s in mailbox %s received at %s to deletion queue",
// 			parsed["subject"], parsed["message_id"], mbox.info.Name, receivedAt)
// 		toDel = append(toDel, msg.Uid)

// 	}

// 	return nil
// }

func (mbox *Mailbox) DeleteMessages(
	// msg *imap.Message,
	uid uint32,
	c *client.Client,
) error {

	var (
		deleteFlagItem = imap.FormatFlagsOp(imap.AddFlags, true)
		deleteFlag     = []interface{}{imap.DeletedFlag}
	)
	seqset := new(imap.SeqSet)
	// seqset.AddNum(msg.Uid)
	seqset.AddNum(uid)
	err := c.UidStore(seqset, deleteFlagItem, deleteFlag, nil)

	if err != nil {
		return fmt.Errorf("mark as deleted failed: %w", err)
	}

	return nil
}

// func (a ActionMove) Perform(msg *imap.Message, c *client.Client) error {
// 	seqset := new(imap.SeqSet)
// 	seqset.AddNum(msg.Uid)

// 	err := c.UidMove(seqset, a.dest)

// 	if err != nil {
// 		return fmt.Errorf("move message failed: %w", err)
// 	}

// 	return nil
// }

func (mbox *Mailbox) MoveMessages(
	uid uint32,
	c *client.Client,
) error {

	/*
		seqset1 := new(imap.SeqSet)
		seqset1.AddNum(uids...)

		err := mbox.client.UidMove(seqset1, "z-archive")
		if err != nil {
			log.Err(err).Msg("archive of messages failed with error")
		}

		for i, mid := range toDel {
			log.Info().Msgf("deleting message %d with id %s in mailbox %s",
				i, mid, mbox.info.Name)
			// if err := DeleteMessage(mid, mbox.client); err != nil {
			// 	log.Err(err).Msgf("deletion of message with ID %d failed with error", mid)
			// }

			if err := MoveMessage(mid, mbox.client); err != nil {
				log.Err(err).Msgf("archive of message with ID %d failed with error", mid)
			}
		}

	*/

	seqset := new(imap.SeqSet)
	seqset.AddNum(uid)

	err := c.UidMove(seqset, "z-archive")

	if err != nil {
		return fmt.Errorf("move message failed: %w", err)
	}

	return nil
}

func (mbox *Mailbox) getMessageIdsInWindow(start, end time.Time) []uint32 {
	criteria := imap.NewSearchCriteria()
	if !start.IsZero() {
		criteria.Since = start
	}
	if !end.IsZero() {
		criteria.Before = end
	}

	uids, err := mbox.client.Search(criteria)
	if err != nil {
		log.Err(err).Msgf("Unable to search for messages in mailbox %s", mbox.info.Name)
		return []uint32{}
	}
	log.Info().Msgf(
		"Found %d new messages in mailbox %s for start %s and end %s.",
		len(uids),
		mbox.info.Name,
		start.String(),
		end.String(),
	)
	return uids
}

func (mbox *Mailbox) Cleanup() {
	if err := mbox.client.Expunge(nil); err != nil {
		log.Err(err).Msgf("Mailbox cleanup for mailbox %s failed with error", mbox.info.Name)
	}
}

/*


//MarkMsgSeen marking message as seen on the server
func (client *Client) MarkMsgSeen(msg *Message) {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(msg.UID)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	err := client.Imap.UidStore(seqSet, item, flags, nil)
	if err != nil {
		log.Fatal(err)
	}
}

{
	// Let's assume c is a client
var c *client.Client

// Select INBOX
_, err := c.Select("INBOX", false)
if err != nil {
    log.Fatal(err)
}

// Set search criteria
criteria := imap.NewSearchCriteria()
criteria.WithoutFlags = []string{imap.SeenFlag}
ids, err := c.Search(criteria)
if err != nil {
    log.Fatal(err)
}
log.Println("IDs found:", ids)

if len(ids) > 0 {
    seqset := new(imap.SeqSet)
    seqset.AddNum(ids...)

    messages := make(chan *imap.Message, 10)
    done := make(chan error, 1)
    go func() {
        done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
    }()

    log.Println("Unseen messages:")
    for msg := range messages {
        log.Println("* " + msg.Envelope.Subject)
    }

    if err := <-done; err != nil {
        log.Fatal(err)
    }
}

log.Println("Done!")
}

// WaitNewMsgs ...
func (client *Client) WaitNewMsgs(msgs chan<- *Message, pollInterval time.Duration) {
	idleClient := idle.NewClient(client.Imap)

	updates := make(chan imapClient.Update)
	client.Imap.Updates = updates

	done := make(chan error, 1)
	go func() {
		done <- idleClient.IdleWithFallback(nil, pollInterval)
	}()

	for {
		select {
		case update := <-updates:
			_, ok := update.(*imapClient.MailboxUpdate)
			if ok {
				log.Println("Got Mailbox update")
				for _, msg := range client.FetchUnseenMsgs() {
					msgs <- msg
				}
			}
		case err := <-done:
			if err != nil {
				log.Fatal(err)
			}
			log.Println("No idling anymore")
			return
		}
	}

}

*/

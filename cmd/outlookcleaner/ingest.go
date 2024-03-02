package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/ozgio/strutil"
	"github.com/rs/zerolog/log"
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

func Save(folder string, messages *[]Message) {
	file, _ := json.MarshalIndent(messages, "", " ")

	_ = os.WriteFile("./emails/"+strutil.ToSnakeCase(folder)+".json", file, 0600)
}

func moveReceipts(sourceFolder, destFolder string) error {
	return nil
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

func foo() {
	client := &http.Client{}
	formData := map[string][]string{
		"fields": {"emails.address"},
	}
	reqBody, _ := json.Marshal(formData)
	req, err := http.NewRequest(
		"GET",
		"https://api.data-axle.com/v1/people/search",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("API status: %d\n", resp.StatusCode)
	fmt.Printf("API Response as struct %+v\n", string(bodyBytes))
}

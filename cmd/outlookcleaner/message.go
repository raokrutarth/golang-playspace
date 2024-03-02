package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"github.com/ozgio/strutil"
	"github.com/rs/zerolog/log"
)

type Message map[string]string

type MessageBody struct {
	MIMEType string
	Message  string
}

func ParseMessageBody(msg *imap.Message) []*MessageBody {
	// Get the whole message body
	var section imap.BodySectionName

	r := msg.GetBody(&section)
	if r == nil {
		log.Error().Msg("message body is empty")
		return nil
	}

	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Err(err).Msg("failed to create message reader")
		return nil
	}

	// Process each message's part
	output := []*MessageBody{}
	var p *mail.Part
	i := 0
	for {
		p, err = mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			log.Err(err).Msgf("failed to read message part %d", i)
		}
		i++

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			// This is the message's text (can be plain-text or HTML)
			b, _ := ioutil.ReadAll(p.Body)
			log.Info().Msgf("Get message content of length %d", len(b))
			output = append(output, &MessageBody{
				MIMEType: "",
				Message:  string(b),
			})
		default:
			log.Info().Msgf("Ignoring body section of type %T", h)
		}
	}
	return output
}

func ParseMessage(imapMsg *imap.Message) Message {

	parsed := map[string]string{}
	parsed["uid"] = strconv.FormatUint(uint64(imapMsg.Uid), 10)
	parsed["subject"] = imapMsg.Envelope.Subject
	parsed["receive_at"] = fmt.Sprint(imapMsg.InternalDate)
	parsed["from"] = fmt.Sprint(imapMsg.Envelope.From[0])
	parsed["to"] = fmt.Sprint(imapMsg.Envelope.To[0])
	parsed["date"] = fmt.Sprint(imapMsg.Envelope.Date)
	parsed["message_id"] = fmt.Sprint(imapMsg.Envelope.MessageId)
	// parsed["body_lit"] = strutil.Summary(fmt.Sprint(imapMsg.GetBody()), 50, "...")

	var len int
	for i, value := range imapMsg.Body {
		len = value.Len()
		buf := make([]byte, len)
		n, err := value.Read(buf)
		if err != nil {
			log.Err(err).Msg("Failed to read message value with error.")
			continue
		}
		if n != len {
			log.Error().Msg("Didn't read correct length")
		}
		parsed[fmt.Sprintf("body_%v", i)] = strutil.Summary(string(buf), 500, "...")
		log.Info().Msgf("body value for subject %s: %v", imapMsg.Envelope.Subject, value)
	}
	return parsed
}

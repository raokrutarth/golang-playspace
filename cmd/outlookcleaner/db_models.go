package main

import (
	"database/sql"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	MessageID       string `gorm:"unique"`
	From            string `gorm:"size:255,index"`
	FromName        string
	To              string `gorm:"size:255"`
	Body            string
	Subject         string
	ReceivedAt      time.Time
	RemoteDeletedAt time.Time
	OpenedAt        sql.NullTime
	MailBoxFolder   string
	SizeBytes       uint32
	IsSeen          bool
	IsFlagged       bool
	IsReceipt       bool
	AttachmentNames string
	Attributes      datatypes.JSON
}
